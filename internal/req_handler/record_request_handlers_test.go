package req_handler

import (
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"testing"

	"github.com/multimario_api/internal/db"
	"github.com/multimario_api/internal/repository/races"
	testutils "github.com/multimario_api/internal/testing"
	"github.com/multimario_api/internal/twitch"
	_ "github.com/ncruces/go-sqlite3/driver"
)

/*
* Tests for race request handlers
 */

//Race specific test struct
type recordTestDB struct {
	//TODO: This might be weird naming since this is mostly a wrapper for testutils.TestDB, but it's nbd
	testDB testutils.TestDB
	recordIDs []int64
	playerIDs []int64
	raceIDs []int64
	runIDs map[int64][]int64
}

//Races to add
type raceStruct struct {
	Category string
	Date string
	Status string
	StartTime string
}

//Players to add
type playerStruct struct {
	DisplayName string
	TwitchName string
}

type recordStruct struct {
	RaceIDIndex int
	PlayerIndex int
	GameTimes map[string]string //{gamecategory: time}
	Estimates map[string]string //{gamecategory: estimate}
}

//Test Add Record
func TestCreateRecord(t *testing.T) {
	//Get test DB and handler
	tdb := initRecordHandlerTestDB(t)
	h := &ReqHandler{tdb.testDB.Database}

	tests := recordsGetPOSTTests()
	for _, test := range tests {
		testutils.CallMutationHandler(t, test, h.CreateRecord)
	}
}

//Test UpdateRecord
func TestUpdateRecord(t *testing.T) {
	//Get test DB and handler
	tdb := initRecordHandlerTestDB(t)
	h := &ReqHandler{tdb.testDB.Database}

	tests := recordsGetPATCHTests()
	for _, test := range tests {
		testutils.CallMutationHandler(t, test, h.UpdateRecord)
	}
}

//Test GetRecords
func TestGetRecords(t *testing.T) {
	//Get teset DB and handler
	tdb := initRecordHandlerTestDB(t)
	h := &ReqHandler{tdb.testDB.Database}

	tests := recordsGetGETTests()

	for _, test := range tests {
		res := testutils.CallQueryHandler(t, test, h.GetRaceRecords)

		//If test failed, body doesn't have all this information
		resSuccess, ok := res["success"].(bool)
		if !ok {
			t.Fatalf("%s: unable to parse success field as boolean", test.TestName)
		}
		if !resSuccess {
			continue
		}

		//Confirm return type
		playerArr, ok := res["records"].([]any)
		if !ok {
			t.Errorf("%s: unable to parse records as array", test.TestName)
		}

		for _, a := range playerArr {
			obj, ok := a.(map[string]any)
			if !ok {
				t.Errorf("%s: unable to parse records objects", test.TestName)
				continue
			}

			//Validate object types
			_, ok = obj["player_name"].(string)
			if !ok {
				t.Errorf("%s: unable to parse player name as string", test.TestName)
				continue
			}
			raceID, ok := obj["race_id"].(float64)
			if !ok {
				t.Errorf("%s: unable to parse race id as int", test.TestName)
				continue
			}

			//Make sure they match the params
			if len(test.URLParams["race_id"]) > 0 {
				if !slices.Contains(test.URLParams["race_id"], strconv.Itoa(int(raceID))) {
					t.Errorf("%s: race_id not filtered", test.TestName)
					continue
				}
			}

			//TODO: Test the rest of the parameters
		}
	}
}

//Create test DB specific to these handlers
func initRecordHandlerTestDB(t *testing.T) recordTestDB {
	t.Helper()

	//Create test DB
	tdb := testutils.CreateTestDB(t)

	races := []raceStruct {
		{"602", "1000-07-13", "upcoming", "9:00:00"},
		{"246", "3214-05-16", "upcoming", "9:00:00"},
		{"sandbox_any%", "1002-10-03", "completed", "9:00:00"},
		{"1862", "3054-12-25", "upcoming", "02:00:00"},
		{"sandbox_100%", "4123-01-31", "upcoming", "9:00:00"},
		{"540", "2025-04-01", "in_progress", "11:00:00"},
		{"602", "2020-12-12", "upcoming", "3:00:00"},
		{"1862", "0000-12-25", "upcoming", "30:00:00"},
	}

	//Valid twitch names for twitch client
	validNames := []string {
		"odme_",
		"expreli_",
		"thorn_97",
		"fizz64",
		"zgamut",
		"clairdss",
		"jukatox",
		"galax_v",
	}
	testutils.SetMockTwitchClient(validNames)

	players := []playerStruct {
		{"Odme", "odme_"},
		{"", "expreli_"},
		{"jake", "thorn_97"},
		{"fizz", "fizz64"},
		{"", "zgamut"},
		{"me", "clairdss"},
		{"juka", "jukatox"},
		{"galax", "galax_v"},
	}

	records := []recordStruct{
		{0, 0, map[string]string {"sm64_120": "2:00:00", "smg1_120": "6:00:00", "sms_120": "2:55:00", "smg2_242": "9:10:00",}, 
			map[string]string{"sm64_120": "1:51:00", "smg1_120": "5:11:32", "sms_120": "2:59:12", "smg2_242": "8:55:00",}}, 
		{1, 0, map[string]string {"sm64_120": "2:10:00", "smg1_120": "6:50:00", "sms_120": "3:55:00", "smg2_242": "9:15:00",},
			map[string]string{"sm64_120": "1:54:30", "smg1_120": "5:31:32", "sms_120": "3:59:12", "smg2_242": "9:35:30",}},
		{2, 0, map[string]string {"sm64_120": "1:05:30", "smg1_120": "1:40:00", "sms_120": "7:55:00", "smg2_242": "4:20:50",}, 
			map[string]string{"sm64_120": "3:51:00", "smg1_120": "5:21:32", "sms_120": "3:59:32", "smg2_242": "8:45:06",}},
		{0, 2, map[string]string {"sm64_70": "2:00:00", "smo_any%": "0:43:54", "sms_any%": "1:43:31",},
			map[string]string{"sm64_70": "1:50:00", "smo_any%": "1:43:54", "sms_any%": "4:13:31",}}, 
		{5, 5, map[string]string { "sm64_70": "2:10:00", "smo_any%": "1:23:54", "sms_any%": "1:45:34", "smg1_any%": "2:43:12", "smg2_any%": "3:12:43", "sm3dw_any%": "1:43:12", }, 
			map[string]string{"sm64_70": "2:10:00", "smo_any%": "1:23:54", "sms_any%": "1:45:34", "smg1_any%": "2:43:12", "smg2_any%": "3:12:43", "sm3dw_any%": "1:39:09", }}, 
		{1, 7, map[string]string {"sm64_120": "3:10:00", "smo_all_moons": "10:23:54", "sms_120": "5:45:34", "smg1_120": "2:13:12", "smg2_242": "6:22:73", "sm3dw_380": "4:43:12",}, 
			map[string]string{"sm64_120": "3:14:00", "smo_all_moons": "9:13:54", "sms_120": "5:45:34", "smg1_120": "7:17:12", "smg2_242": "1:32:73", "sm3dw_380": "8:33:12",}}, 
		{6, 6, map[string]string {"sm64_120": "3:10:00", "sms_120": "5:45:34", "smg1_120": "2:13:12", "smg2_242": "6:22:73",}, 
			map[string]string{"sm64_120": "3:14:00", "sms_120": "5:45:34", "smg1_120": "7:17:12", "smg2_242": "1:32:73"}}, 
		{5, 0, map[string]string { "sm64_70": "1:10:00", "smo_any%": "1:24:54", "sms_any%": "1:45:34", "smg1_any%": "2:43:12", "smg2_any%": "3:12:43", "sm3dw_any%": "1:43:12", }, 
			map[string]string{"sm64_70": "2:10:00", "smo_any%": "1:21:55", "sms_any%": "1:15:34", "smg1_any%": "2:43:12", "smg2_any%": "3:12:43", "sm3dw_any%": "1:39:09", }}, 
		{5, 1, map[string]string { "sm64_70": "2:10:00", "smo_any%": "1:23:54", "sms_any%": "2:45:34", "smg1_any%": "2:43:12", "smg2_any%": "3:12:43", "sm3dw_any%": "1:43:12", }, 
			map[string]string{"sm64_70": "2:10:00", "smo_any%": "1:23:54", "sms_any%": "1:35:34", "smg1_any%": "2:43:12", "smg2_any%": "3:12:43", "sm3dw_any%": "1:39:09", }}, 
		{5, 2, map[string]string { "sm64_70": "2:10:00", "smo_any%": "1:13:54", "sms_any%": "1:45:34", "smg1_any%": "2:43:12", "smg2_any%": "3:12:43", "sm3dw_any%": "1:43:12", }, 
			map[string]string{"sm64_70": "2:10:00", "smo_any%": "1:13:51", "sms_any%": "1:45:34", "smg1_any%": "2:43:12", "smg2_any%": "3:12:43", "sm3dw_any%": "1:39:09", }}, 
		{5, 3, map[string]string { "sm64_70": "2:10:00", "smo_any%": "1:23:54", "sms_any%": "1:45:34", "smg1_any%": "2:43:12", "smg2_any%": "3:12:43", "sm3dw_any%": "1:43:12", }, 
			map[string]string{"sm64_70": "2:10:00", "smo_any%": "1:23:54", "sms_any%": "1:45:34", "smg1_any%": "2:43:12", "smg2_any%": "3:12:43", "sm3dw_any%": "1:39:09", }}, 

	}

	return populateTestDB(t, tdb, players, races, records)
}

//Helper function to handle SQL calls to the test database
func populateTestDB(t *testing.T, tdb testutils.TestDB, players []playerStruct, raceStruct []raceStruct, records []recordStruct) recordTestDB {
	t.Helper()

	raceIDs := make([]int64, 0)
	for _, r := range raceStruct {
		stmt := fmt.Sprintf("INSERT INTO %s (%s, %s, %s, %s) VALUES (?, ?, ?, ?)",
			db.TableRaces, db.ColRaceRaceCategoryID, db.ColRaceDate, db.ColRaceStatus, db.ColRaceStartTime)
		res, err := tdb.Database.Exec(stmt, tdb.RaceCatIDs[r.Category], r.Date, r.Status, r.StartTime)
		if err != nil {
			t.Fatalf("unable to init test database: %v", err)
		}

		newRaceID, err := res.LastInsertId()
		if err != nil {
			t.Fatalf("unable to init test database: %v", err)
		}

		raceIDs = append(raceIDs, newRaceID)
	}

	//Init current race
	err := races.InitCurrentRace(tdb.Database)
	if err != nil {
		t.Fatal(err)
	}

	playerIDs := make([]int64, 0)
	for _, p := range players {
		stmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (?)", db.TablePlayers, db.ColPlayerName)
		
		name := p.DisplayName
		if p.DisplayName == "" {
			name = p.TwitchName
		}

		res, err := tdb.Database.Exec(stmt, name)
		if err != nil {
			t.Fatalf("unable to init test database: %v", err)
		}

		newPlayerID, err := res.LastInsertId()
		if err != nil {
			t.Fatalf("unable to init test database: %v", err)
		}

		playerIDs = append(playerIDs, newPlayerID)

		//Add the socials table
		stmt = fmt.Sprintf("INSERT INTO %s (%s, %s, %s) VALUES (?, ?, ?)", db.TableSocials, db.ColSocialsPlayerID, db.ColSocialsPlatformName, db.ColSocialsPlatformUserID)
		pTwitchID, err := twitch.Client.GetTwitchIDFromName(p.TwitchName)
		if err != nil {
			t.Fatalf("unable to get player twitch id")
		}

		res, err = tdb.Database.Exec(stmt, newPlayerID, "twitch", pTwitchID)
		if err != nil {
			t.Fatalf("unable to add player's twitch information: %s", err.Error())
		}
	}

	recordIDs := make([]int64, 0)
	runIDs := make(map[int64][]int64) //{recordID: runIDs}
	for _, r := range records {
		//Add record
		stmt := fmt.Sprintf("INSERT INTO %s (%s, %s, %s, %s) VALUES (?, ?, ?, ?)", 
			db.TableRecords, db.ColRecordsRaceID, db.ColRecordsPlayerID, db.ColRecordsFinishTime, db.ColRecordsNumCollected)

		res, err := tdb.Database.Exec(stmt, raceIDs[r.RaceIDIndex], playerIDs[r.PlayerIndex], "20:00:00", 15)
		if err != nil {
			t.Fatalf("unable to init test database: %v", err)
		}

		newRecID, err:= res.LastInsertId()
		if err != nil {
			t.Fatalf("unable to init test database: %v", err)
		}
		recordIDs = append(recordIDs, newRecID)

		//Add runs
		newRunIds := make([]int64, 0)
		for catName, time := range r.GameTimes {
			stmt := fmt.Sprintf("INSERT INTO %s (%s, %s, %s, %s, %s) VALUES (?, ?, ?, ?, ?)",
				db.TableRuns, db.ColRunRaceRecordID, db.ColRunGameCategoryID, db.ColRunTime, db.ColRunEstimate, db.ColRunNum)

			res, err := tdb.Database.Exec(stmt, newRecID, tdb.GameCatIDs[catName], time, r.Estimates[catName], 0)
			if err != nil {
				t.Fatalf("unable to init test database: %v", err)
			}

			newRunID, err := res.LastInsertId()
			if err != nil {
				t.Fatalf("unable to init test database: %v", err)
			}

			newRunIds = append(newRunIds, newRunID)
		}
		runIDs[newRecID] = newRunIds
	}

	//Get the thing
	return recordTestDB{
		testDB: tdb,
		recordIDs: recordIDs,
		playerIDs: playerIDs, 
		raceIDs: raceIDs,
		runIDs: runIDs,
	}
}

//Get Record Post Tests
func recordsGetPOSTTests() []testutils.MutationHandlerTest {
	return []testutils.MutationHandlerTest{
		{
		TestName: "ValidAllFields",
		Body: map[string]any {
			"race_id": 7,
			"player_name": "Odme",
			"finish_time": "20:00:00",
			"num_collected": 602,
			"runs": []map[string]any{
				{
					"category_name": "sm64_120",
					"time": "2:10:00",
					"estimate": "1:50:00",
				}, {
					"category_name": "smg1_120",
					"time": "5:51:00",
					"estimate": "5:00:00",
				}, {
					"category_name": "sms_120",
					"time": "3:00:00",
					"estimate": "2:51:51",
				}, {
					"category_name": "smg2_242",
					"time": "9:00:00",
					"estimate": "8:59:59",
				},
			},
		},
		RequestType: "POST",
		Pattern: "POST /records",
		URL: "/records",
		ExpectedResponseCode: http.StatusOK,
		ExpectedSuccess: true,
		}, {
		TestName: "ValidNoFinishTime",
		Body: map[string]any {
			"race_id": 3,
			"player_name": "expreli_",
			"num_collected":12,
			"runs": []map[string]any{
				{
					"category_name": "sm64_70",
					"time": "1:10:00",
					"estimate": "0:50:00",
				}, {
					"category_name": "sms_any%",
					"time": "1:51:00",
					"estimate": "1:20:00",
				}, {
					"category_name": "smo_any%",
					"time": "1:00:00",
					"estimate": "0:51:51",
				},
			},
		},
		RequestType: "POST",
		Pattern: "POST /records",
		URL: "/records",
		ExpectedResponseCode: http.StatusOK,
		ExpectedSuccess: true,
		}, {
		TestName: "ValidNoCollectibles",
		Body: map[string]any {
			"race_id": 3,
			"player_name": "me",
			"runs": []map[string]any{
				{
					"category_name": "sm64_70",
					"time": "1:10:00",
					"estimate": "0:50:00",
				}, {
					"category_name": "sms_any%",
					"time": "1:51:00",
					"estimate": "1:20:00",
				}, {
					"category_name": "smo_any%",
					"time": "1:00:00",
					"estimate": "0:51:51",
				},
			},
		},
		RequestType: "POST",
		Pattern: "POST /records",
		URL: "/records",
		ExpectedResponseCode: http.StatusOK,
		ExpectedSuccess: true,
		}, {
		TestName: "ValidNoTwitchName",
		Body: map[string]any {
			"race_id": 5,
			"player_name": "clairdss",
			"num_collected":12,
			"runs": []map[string]any{
				{
					"category_name": "sm64_120",
					"time": "2:10:00",
					"estimate": "1:50:00",
				}, {
					"category_name": "sms_120",
					"time": "3:51:00",
					"estimate": "3:20:00",
				}, {
					"category_name": "smo_all_moons",
					"time": "7:00:00",
					"estimate": "6:51:51",
				},
			},
		},
		RequestType: "POST",
		Pattern: "POST /records",
		URL: "/records",
		ExpectedResponseCode: http.StatusOK,
		ExpectedSuccess: true,
		}, {
		TestName: "ValidNoRuns",
		Body: map[string]any {
			"race_id": 4,
			"player_name": "zgamut",
			"runs": make([]map[string]any, 0),
		},
		RequestType: "POST",
		Pattern: "POST /records",
		URL: "/records",
		ExpectedResponseCode: http.StatusOK,
		ExpectedSuccess: true,
		}, {
		TestName: "InvalidBadEstimateFormat",
		Body: map[string]any {
			"race_id": 3,
			"player_name": "jukatox",
			"runs": []map[string]any{
				{
					"category_name": "sm64_70",
					"time": "1:10:00",
					"estimate": "50:00",
				}, {
					"category_name": "sms_any%",
					"time": "1:51:00",
					"estimate": "1:20:00",
				}, {
					"category_name": "smo_any%",
					"time": "1:00:00",
					"estimate": "0:51:51",
				},
			},
		},
		RequestType: "POST",
		Pattern: "POST /records",
		URL: "/records",
		ExpectedResponseCode: http.StatusBadRequest,
		ExpectedSuccess: false,
		}, {
		TestName: "InvalidBadGameTimeFormat",
		Body: map[string]any {
			"race_id": 3,
			"player_name": "jukatox",
			"runs": []map[string]any{
				{
					"category_name": "sm64_70",
					"time": "1:10:00",
					"estimate": "00:50:00",
				}, {
					"category_name": "sms_any%",
					"time": "1:51",
					"estimate": "1:20:00",
				}, {
					"category_name": "smo_any%",
					"time": "1:00:00",
					"estimate": "0:51:51",
				},
			},
		},
		RequestType: "POST",
		Pattern: "POST /records",
		URL: "/records",
		ExpectedResponseCode: http.StatusBadRequest,
		ExpectedSuccess: false,
		}, {
		TestName: "InvalidRaceDoesntExist",
		Body: map[string]any {
			"race_id": 100,
			"player_name": "jukatox",
			"runs": []map[string]any{
				{
					"category_name": "sm64_70",
					"time": "1:10:00",
					"estimate": "00:50:00",
				}, {
					"category_name": "sms_any%",
					"time": "1:51:00",
					"estimate": "1:20:00",
				}, {
					"category_name": "smo_any%",
					"time": "1:00:00",
					"estimate": "0:51:51",
				},
			},
		},
		RequestType: "POST",
		Pattern: "POST /records",
		URL: "/records",
		ExpectedResponseCode: http.StatusBadRequest,
		ExpectedSuccess: false,
		}, {
		TestName: "InvalidPlayerDoesntExist",
		Body: map[string]any {
			"race_id": 2,
			"player_name": "invalid12",
			"runs": []map[string]any{
				{
					"category_name": "sm64_70",
					"time": "1:10:00",
					"estimate": "00:50:00",
				}, {
					"category_name": "sms_any%",
					"time": "1:51:00",
					"estimate": "1:20:00",
				}, {
					"category_name": "smg1_any%",
					"time": "1:00:00",
					"estimate": "0:51:51",
				}, {
					"category_name": "smg2_any%",
					"time": "3:00:00",
					"estimate": "2:51:51",
				},
			},
		},
		RequestType: "POST",
		Pattern: "POST /records",
		URL: "/records",
		ExpectedResponseCode: http.StatusBadRequest,
		ExpectedSuccess: false,
		}, {
		TestName: "InvalidCategoryNotInRace",
		Body: map[string]any {
			"race_id": 3,
			"player_name": "jukatox",
			"runs": []map[string]any{
				{
					"category_name": "sm64_70",
					"time": "1:10:00",
					"estimate": "00:50:00",
				}, {
					"category_name": "sms_any%",
					"time": "1:51:00",
					"estimate": "1:20:00",
				}, {
					"category_name": "smo_100%",
					"time": "1:00:00",
					"estimate": "0:51:51",
				},
			},
		},
		RequestType: "POST",
		Pattern: "POST /records",
		URL: "/records",
		ExpectedResponseCode: http.StatusBadRequest,
		ExpectedSuccess: false,
		}, {
		TestName: "InvalidFinishTimeFormat",
		Body: map[string]any {
			"race_id": 3,
			"player_name": "jukatox",
			"finish_time": "12 hours",
			"runs": []map[string]any{
				{
					"category_name": "sm64_70",
					"time": "1:10:00",
					"estimate": "00:50:00",
				}, {
					"category_name": "sms_any%",
					"time": "1:51:00",
					"estimate": "1:20:00",
				}, {
					"category_name": "smo_100%",
					"time": "1:00:00",
					"estimate": "0:51:51",
				},
			},
		},
		RequestType: "POST",
		Pattern: "POST /records",
		URL: "/records",
		ExpectedResponseCode: http.StatusBadRequest,
		ExpectedSuccess: false,
		}, {
		TestName: "InvalidNoCategoryName",
		Body: map[string]any {
			"race_id": 3,
			"player_name": "jukatox",
			"runs": []map[string]any{
				{
					"category_name": "sm64_70",
					"time": "1:10:00",
					"estimate": "00:50:00",
				}, {
					"category_name": "sms_any%",
					"time": "1:51:00",
					"estimate": "1:20:00",
				}, {
					"time": "1:00:00",
					"estimate": "0:51:51",
				},
			},
		},
		RequestType: "POST",
		Pattern: "POST /records",
		URL: "/records",
		ExpectedResponseCode: http.StatusBadRequest,
		ExpectedSuccess: false,
		}, {
		TestName: "InvalidEmpty",
		Body: make(map[string]any),
		RequestType: "POST",
		Pattern: "POST /records",
		URL: "/records",
		ExpectedResponseCode: http.StatusBadRequest,
		ExpectedSuccess: false,
		}, {
		TestName: "InvalidNoPlayer",
		Body: map[string]any {
			"race_id": 3,
			"runs": []map[string]any{
				{
					"category_name": "sm64_70",
					"time": "1:10:00",
					"estimate": "00:50:00",
				}, {
					"category_name": "sms_any%",
					"time": "1:51:00",
					"estimate": "1:20:00",
				}, {
					"time": "1:00:00",
					"estimate": "0:51:51",
				},
			},
		},
		RequestType: "POST",
		Pattern: "POST /records",
		URL: "/records",
		ExpectedResponseCode: http.StatusBadRequest,
		ExpectedSuccess: false,
		}, {
		TestName: "InvalidEmpty",
		Body: make(map[string]any),
		RequestType: "POST",
		Pattern: "POST /records",
		URL: "/records",
		ExpectedResponseCode: http.StatusBadRequest,
		ExpectedSuccess: false,
		}, {
		TestName: "InvalidRecordExists",
		Body: map[string]any {
			"race_id": 1,
			"player_name": "Odme",
			"runs": []map[string]any{
				{
					"category_name": "sm64_120",
					"time": "2:10:00",
					"estimate": "1:50:00",
				}, {
					"category_name": "smg1_120",
					"time": "5:51:00",
					"estimate": "5:00:00",
				}, {
					"category_name": "sms_120",
					"time": "3:00:00",
					"estimate": "2:51:51",
				}, {
					"category_name": "smg2_242",
					"time": "9:00:00",
					"estimate": "8:59:59",
				},
			},
		},
		RequestType: "POST",
		Pattern: "POST /records",
		URL: "/records",
		ExpectedResponseCode: http.StatusBadRequest,
		ExpectedSuccess: false,
		},
	}
}

//Get records PATCH tests
func recordsGetPATCHTests() []testutils.MutationHandlerTest {
	return []testutils.MutationHandlerTest {
		{
			TestName: "ValidAllFields",
			Body: map[string]any {
				"finish_time": "18:12:13",
				"num_collected": 603,
			},
			RequestType: "PATCH",
			Pattern: "PATCH /records/{race_id}/{player_name}",
			URL: "/records/1/odme_",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidNoNumCollected",
			Body: map[string]any {
				"finish_time": "18:12:13",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /records/{race_id}/{player_name}",
			URL: "/records/6/clairdss",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidNoFinish",
			Body: map[string]any {
				"num_collected": 604,
			},
			RequestType: "PATCH",
			Pattern: "PATCH /records/{race_id}/{player_name}",
			URL: "/records/1/thorn_97",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidEmpty",
			Body: make(map[string]any),
			RequestType: "PATCH",
			Pattern: "PATCH /records/{race_id}/{player_name}",
			URL: "/records/2/galax",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "InvalidPlayer",
			Body: map[string]any {
				"finish_time": "3:12:13",
				"num_collected": 12,
			},
			RequestType: "PATCH",
			Pattern: "PATCH /records/{race_id}/{player_name}",
			URL: "/records/5/invalid",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidRace",
			Body: map[string]any {
				"finish_time": "3:12:13",
				"num_collected": 12,
			},
			RequestType: "PATCH",
			Pattern: "PATCH /records/{race_id}/{player_name}",
			URL: "/records/99/zgamut",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidTimeFormat",
			Body: map[string]any {
				"finish_time": "12:13",
				"num_collected": 13,
			},
			RequestType: "PATCH",
			Pattern: "PATCH /records/{race_id}/{player_name}",
			URL: "/records/1/thorn_97",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidCollectedFormat",
			Body: map[string]any {
				"num_collected": "13",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /records/{race_id}/{player_name}",
			URL: "/records/1/thorn_97",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, 
	}
}

//Get GET tests
func recordsGetGETTests() []testutils.QueryHandlerTest{
	return []testutils.QueryHandlerTest{
		{
			TestName: "ValidAllRecords",
			URLParams: make(map[string][]string),
			Pattern: "GET /records",
			URL: "/records",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true, 
		}, {
			TestName: "ValidFromOdme",
			URLParams: map[string][]string {
				"player_name": {"Odme"},
			},
			Pattern: "GET /records",
			URL: "/records",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true, 
		}, {
			TestName: "ValidFromOdmeOrMe",
			URLParams: map[string][]string {
				"player_name": {"Odme", "me"},
			},
			Pattern: "GET /records",
			URL: "/records",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true, 
		}, {
			TestName: "ValidBefore2020",
			URLParams: map[string][]string {
				"before": {"2020-01-01"},
			},
			Pattern: "GET /records",
			URL: "/records",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true, 
		}, {
			TestName: "ValidBefore2000After1000",
			URLParams: map[string][]string {
				"before": {"2000-01-01"},
				"after": {"1000-12-31"},
			},
			Pattern: "GET /records",
			URL: "/records",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true, 
		}, {
			TestName: "ValidOn",
			URLParams: map[string][]string {
				"on": {"1000-07-13"},
			},
			Pattern: "GET /records",
			URL: "/records",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true, 
		}, {
			TestName: "ValidSlowerThan30hr",
			URLParams: map[string][]string {
				"time_lowerthan": {"30:00:00"},
			},
			Pattern: "GET /records",
			URL: "/records",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true, 
		}, {
			TestName: "ValidSlowerThan20hr",
			URLParams: map[string][]string {
				"time_lowerthan": {"20:00:00"},
			},
			Pattern: "GET /records",
			URL: "/records",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true, 
		}, {
			TestName: "ValidFasterThan20hr",
			URLParams: map[string][]string {
				"time_greaterthan": {"20:00:00"},
			},
			Pattern: "GET /records",
			URL: "/records",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true, 
		}, {
			TestName: "ValidFasterThan10hr",
			URLParams: map[string][]string {
				"time_greaterthan": {"10:00:00"},
			},
			Pattern: "GET /records",
			URL: "/records",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true, 
		}, {
			TestName: "Valid602s",
			URLParams: map[string][]string {
				"category": {"602"},
			},
			Pattern: "GET /records",
			URL: "/records",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true, 
		}, {
			TestName: "ValidFromThorn",
			URLParams: map[string][]string {
				"player_name": {"thorn_97"},
			},
			Pattern: "GET /records",
			URL: "/records",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true, 
		}, {
			TestName: "ValidFromThornAndJake",
			URLParams: map[string][]string {
				"player_name": {"thorn_97", "jake"},
			},
			Pattern: "GET /records",
			URL: "/records",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true, 
		}, {
			TestName: "ValidPlayerDoesntExist",
			URLParams: map[string][]string {
				"player_name": {"invalid"},
			},
			Pattern: "GET /records",
			URL: "/records",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true, 
		}, {
			TestName: "InvalidDate",
			URLParams: map[string][]string {
				"after": {"12/31/1000"},
			},
			Pattern: "GET /records",
			URL: "/records",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false, 
		}, {
			TestName: "InvalidTime",
			URLParams: map[string][]string {
				"time_lowerthan": {"3:69:12"},
			},
			Pattern: "GET /records",
			URL: "/records",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false, 
		}, 
	}
}