package req_handler

import (
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"testing"

	"github.com/multimario_api/internal/db"
	testutils "github.com/multimario_api/internal/testing"
	_ "github.com/ncruces/go-sqlite3/driver"
)

/*
* Tests for race request handlers
 */

//Race specific test struct
type raceTestDB struct {
	//TODO: This might be weird naming since this is mostly a wrapper for testutils.TestDB, but it's nbd
	testDB testutils.TestDB
	raceIDs []int64
}

//Create test DB specific to these handlers
func initRaceHandlerTestDB(t *testing.T) raceTestDB {
	t.Helper()

	//Create test DB
	tdb := testutils.CreateTestDB(t)

	//Races to add
	type raceStruct struct {
		Category string
		Date string
		Status string
		StartTime string
	}

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

	raceIDs := make([]int64, 0)
	for _, r := range races {
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

		raceIDs= append(raceIDs, newRaceID)
	}

	return raceTestDB{
		testDB: tdb,
		raceIDs: raceIDs,
	}
}

//Test CreateRace()
func TestCreateRace(t *testing.T) {
	//Get test DB and handler
	tdb := initRaceHandlerTestDB(t)
	h := &ReqHandler{tdb.testDB.Database}

	tests := getPOSTTests()
	for _, test := range tests {
		res := testutils.CallMutationHandler(t, test, h.CreateRace)
		
		//Validate response
		if res["success"] == true {
			_, ok := res["id"].(float64)
			if !ok {
				t.Errorf("%s: race id could not be parsed as int", test.TestName)
			}
		}
	}
}

//Test UpdateRace
func TestUpdateRace(t *testing.T) {
	//Get test DB and handler
	tdb := initRaceHandlerTestDB(t)
	h := &ReqHandler{tdb.testDB.Database}

	tests := getPATCHTests(tdb.raceIDs)
	for _, test := range tests {
		testutils.CallMutationHandler(t, test, h.UpdateRace)
	}
}

//Test GetRaces
func TestGetRaces(t *testing.T) {
	//Lowkey this function is a mess but its a test so whatever
	//Get teset DB and handler
	tdb := initRaceHandlerTestDB(t)
	h := &ReqHandler{tdb.testDB.Database}

	tests := getGETTests()

	for _, test := range tests {
		res := testutils.CallQueryHandler(t, test, h.GetRaces)

		//Get test bounds
		beforeDate, afterDate := getDateBounds(test.URLParams["before"], test.URLParams["after"])

		//Confirm return type
		raceArr, ok := res["races"].([]any)
		if !ok {
			t.Errorf("%s: unable to parse races as array", test.TestName)
		}

		for _, a := range raceArr {
			obj, ok := a.(map[string]any)
			if !ok {
				t.Errorf("%s: unable to parse race objects", test.TestName)
				continue
			}

			//Validate object types
			raceID, ok := obj["id"].(float64)
			if !ok {
				t.Errorf("%s: unable to parse race id as int", test.TestName)
				continue
			}
			catName, ok := obj["category"].(string)
			if !ok {
				t.Errorf("%s: unable to parse category name as string", test.TestName)
				continue
			}
			date, ok := obj["date"].(string)
			if !ok {
				t.Errorf("%s: unable to parse date as string", test.TestName)
				continue
			}
			status, ok := obj["status"].(string)
			if !ok {
				t.Errorf("%s: unable to parse status as string", test.TestName)
				continue
			}
			_, ok = obj["start_time"].(string)
			if !ok {
				t.Errorf("%s: unable to parse start time as string", test.TestName)
				continue
			}

			//Make sure they match the params
			if len(test.URLParams["race_id"]) > 0 {
				validID := slices.Contains(test.URLParams["race_id"], strconv.Itoa(int(raceID)))
				if !validID {
					t.Errorf("%s: race id not filtered", test.TestName)
					continue
				}
			}

			if len(test.URLParams["category"]) > 0 {
				if !slices.Contains(test.URLParams["category"], catName) {
					t.Errorf("%s: race category not filtered", test.TestName)
					continue
				}
			}

			if len(test.URLParams["status"]) > 0 {
				if !slices.Contains(test.URLParams["status"], status) {
					t.Errorf("%s: race status not filtered", test.TestName)
					continue
				}
			}
			
			if len(test.URLParams["before"]) > 0 {
				if date > beforeDate {
					t.Errorf("%s: date not filtered", test.TestName)
					continue
				}
			}

			if len(test.URLParams["after"]) > 0 {
				if date < afterDate {
					t.Errorf("%s: date not filtered", test.TestName)
					continue
				}
			}

			if len(test.URLParams["on"]) > 0 {
				if !slices.Contains(test.URLParams["on"], date) {
					t.Errorf("%s: date not filtered", test.TestName)
					continue
				}
			}
		}
	}
}

//Helper to confirm that GET dates are within the timeframe
func getDateBounds(befores []string, afters []string) (string, string) {
	//Make sure date is between these
	//Get latest before date, since it encompasses all other dates
	var beforeDate string
	for i, date := range befores {
		if i == 0 || date > beforeDate {
			beforeDate = date
		}
	}

	//Get earliest after date
	var afterDate string
	for i, date := range afters {
		if i == 0 || date < afterDate {
			afterDate = date
		}
	}

	return beforeDate, afterDate
}

//Helper to create the tests for the post handler
func getPOSTTests() []testutils.MutationHandlerTest {
	//Create tests
	return []testutils.MutationHandlerTest{{
		//Valid tests
		TestName: "ValidAllFields",
		Body: map[string]any {
			"category": "602",
			"date": "2000-11-16",
			"status": "completed",
			"start_time": "09:00:00",
		},
		RequestType: "POST",
		Pattern: "POST /races",
		URL: "/races",
		ExpectedResponseCode: http.StatusOK,
		ExpectedSuccess: true,
	}, {
		TestName: "ValidNoDate",
		Body: map[string]any {
			"category": "sandbox_any%",
			"status": "upcoming",
			"start_time": "10:00:00",
		},
		RequestType: "POST",
		Pattern: "POST /races",
		URL: "/races",
		ExpectedResponseCode: http.StatusOK,
		ExpectedSuccess: true,
	}, {
		TestName: "ValidEmptyDate",
		Body: map[string]any {
			"category": "sandbox_100%",
			"date": "",
			"status": "completed",
			"start_time": "13:00:00",
		},
		RequestType: "POST",
		Pattern: "POST /races",
		URL: "/races",
		ExpectedResponseCode: http.StatusOK,
		ExpectedSuccess: true,
	}, {
		TestName: "ValidNoStatus",
		Body: map[string]any {
			"category": "1862",
			"date": "1776-07-04",
			"start_time": "11:00:00",
		},
		RequestType: "POST",
		Pattern: "POST /races",
		URL: "/races",
		ExpectedResponseCode: http.StatusOK,
		ExpectedSuccess: true,
	}, {
		TestName: "ValidEmptyStatus",
		Body: map[string]any {
			"category": "246",
			"date": "0001-12-25",
			"status": "",
			"start_time": "15:00:00",
		},
		RequestType: "POST",
		Pattern: "POST /races",
		URL: "/races",
		ExpectedResponseCode: http.StatusOK,
		ExpectedSuccess: true,
	}, {
		TestName: "ValidNoStartTime",
		Body: map[string]any {
			"category": "602",
			"date": "1234-01-23",
			"status": "upcoming",
		},
		RequestType: "POST",
		Pattern: "POST /races",
		URL: "/races",
		ExpectedResponseCode: http.StatusOK,
		ExpectedSuccess: true,
	}, {
		TestName: "ValidEmptyStartTime",
		Body: map[string]any {
			"category": "602",
			"date": "5421-05-14",
			"status": "in_progress",
			"start_time": "",
		},
		RequestType: "POST",
		Pattern: "POST /races",
		URL: "/races",
		ExpectedResponseCode: http.StatusOK,
		ExpectedSuccess: true,
	}, {
		TestName: "ValidNoDateNoStatus",
		Body: map[string]any {
			"category": "sandbox_any%",
			"start_time": "15:12:09",
		}, 
		RequestType: "POST",
		Pattern: "POST /races",
		URL: "/races",
		ExpectedResponseCode: http.StatusOK,
		ExpectedSuccess: true,
	}, {
		TestName: "ValidNoDateNoStartTime",
		Body: map[string]any {
			"category": "sandbox_100%",
			"status": "upcoming",
		},
		RequestType: "POST",
		Pattern: "POST /races",
		URL: "/races",
		ExpectedResponseCode: http.StatusOK,
		ExpectedSuccess: true,
	}, {
		TestName: "ValidNoStatusNoStartTime",
		Body: map[string]any {
			"category": "602",
			"date": "2026-05-14",
		},
		RequestType: "POST",
		Pattern: "POST /races",
		URL: "/races",
		ExpectedResponseCode: http.StatusOK,
		ExpectedSuccess: true,
	}, {
		TestName: "ValidOnlyCategory",
		Body: map[string]any {
			"category": "246",
		},
		RequestType: "POST",
		Pattern: "POST /races",
		URL: "/races",
		ExpectedResponseCode: http.StatusOK,
		ExpectedSuccess: true,
	}, 

	//Invalid tests
	{
		TestName: "InvalidCategoryDoesntExist",
		Body: map[string]any {
			"category": "bad",
			"date": "2000-01-01",
			"status": "completed",
			"start_time": "09:00:00",
		},
		RequestType: "POST",
		Pattern: "POST /races",
		URL: "/races",
		ExpectedResponseCode: http.StatusBadRequest,
		ExpectedSuccess: false,
	}, {
		TestName: "InvalidCategoryEmpty",
		Body: map[string]any {
			"date": "2000-04-12",
			"status": "upcoming",
			"start_time": "09:01:13",
		},
		RequestType: "POST",
		Pattern: "POST /races",
		URL: "/races",
		ExpectedResponseCode: http.StatusBadRequest,
		ExpectedSuccess: false,
	}, {
		TestName: "InvalidDateWrongFormat",
		Body: map[string]any {
			"category": "540",
			"date": "12-13-2054",
			"status": "upcoming",
			"start_time": "09:12:12",
		},
		RequestType: "POST",
		Pattern: "POST /races",
		URL: "/races",
		ExpectedResponseCode: http.StatusBadRequest,
		ExpectedSuccess: false,
	}, {
		TestName: "InvalidStatus",
		Body: map[string]any {
			"category": "sandbox_any%",
			"date": "1234-03-12",
			"status": "idk",
			"start_time": "09:00:00",
		},
		RequestType: "POST",
		Pattern: "POST /races",
		URL: "/races",
		ExpectedResponseCode: http.StatusBadRequest,
		ExpectedSuccess: false,
	}, {
		TestName: "InvalidTime",
		Body: map[string]any {
			"category": "602",
			"date": "1999-01-01",
			"status": "in_progress",
			"start_time": "9am",
		},
		RequestType: "POST",
		Pattern: "POST /races",
		URL: "/races",
		ExpectedResponseCode: http.StatusBadRequest,
		ExpectedSuccess: false,
	}}
}

//Helper to create tests for PATCH requests
func getPATCHTests(raceIDs []int64) []testutils.MutationHandlerTest {
	tests := make([]testutils.MutationHandlerTest, 0)

	//Create tests for each race ID
	for _, id := range raceIDs {
		idTests := []testutils.MutationHandlerTest{{
			TestName: fmt.Sprintf("ID%vValidUpdateAll", id),
			Body: map[string]any {
				"date": "0000-11-16",
				"status": "in_progress",
				"start_time": "11:00:00",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /races/{race_id}",
			URL: fmt.Sprintf("/races/%v", id),
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: fmt.Sprintf("ID%vInvalidDate", id),
			Body: map[string]any {
				"date": "9-15-1222",
				"status": "upcoming",
				"start_time": "11:13:12",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /races/{race_id}",
			URL: fmt.Sprintf("/races/%v", id),
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: fmt.Sprintf("ID%vValidNoStatus", id),
			Body: map[string]any {
				"status": "completed",
				"start_time": "11:00:00.01",
			}, 
			RequestType: "PATCH",
			Pattern: "PATCH /races/{race_id}",
			URL: fmt.Sprintf("/races/%v", id),
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: fmt.Sprintf("ID%vValidEmpty", id),
			Body: make(map[string]any),
			RequestType: "PATCH",
			Pattern: "PATCH /races/{race_id}",
			URL: fmt.Sprintf("/races/%v", id),
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: fmt.Sprintf("ID%vInvalidStartTime", id),
			Body: map[string]any {
				"start_time": "11am",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /races/{race_id}",
			URL: fmt.Sprintf("/races/%v", id),
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}}
		tests = append(tests, idTests...)
	}
	return tests
}

//Returns GET tests
func getGETTests() []testutils.QueryHandlerTest {
	return []testutils.QueryHandlerTest{
		{
			TestName: "ValidNoQueries",
			URLParams: make(map[string][]string),
			Pattern: "GET /races",
			URL: "/races",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidOnly602",
			URLParams: map[string][]string{
				"category": {"602"},
			},
			Pattern: "GET /races",
			URL: "/races",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "Valid602AndSandboxAny%",
			URLParams: map[string][]string {
				"category": {"602", "sandbox_any%"}, 
			},
			Pattern: "GET /races",
			URL: "/races",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidBefore2020",
			URLParams: map[string][]string {
				"before": {"2020-01-01"},
			},
			Pattern: "GET /races",
			URL: "/races",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidBetween1000And2000",
			URLParams: map[string][]string {
				"before": {"2000-01-01"},
				"after": {"1000-12-31"},
			},
			Pattern: "GET /races",
			URL: "/races",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidBetween1000And2000AndBirthOfChrist",
			URLParams: map[string][]string {
				"before": {"2000-01-01"},
				"after": {"1000-12-31"},
				"on": {"0000-12-25"},
			},
			Pattern: "GET /races",
			URL: "/races",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidOnBirthOfChrist",
			URLParams: map[string][]string {
				"on": {"0000-12-25"},
			},
			Pattern: "GET /races",
			URL: "/races",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidBefore4000OrBefore3000",
			URLParams: map[string][]string {
				"before": {"4000-01-01", "3000-01-01"},
			},
			Pattern: "GET /races",
			URL: "/races",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		},
		{
			TestName: "ValidRaces123",
			URLParams: map[string][]string {
				"race_id": {"1", "2", "3"},
			},
			Pattern: "GET /races",
			URL: "/races",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		},
	}
}