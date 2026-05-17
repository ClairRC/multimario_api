package req_handler

import (
	"fmt"
	"net/http"
	"slices"
	"testing"

	"github.com/multimario_api/internal/db"
	testutils "github.com/multimario_api/internal/testing"
	"github.com/multimario_api/internal/twitch"
)

//Tests for player request handlers

//Player Test DB
//Race specific test struct
type playerTestDB struct {
	//TODO: This might be weird naming since this is mostly a wrapper for testutils.TestDB, but it's nbd
	testDB testutils.TestDB
	playerIDs []int64
}

//Create test DB specific to these handlers
func initPlayerHandlerTestDB(t *testing.T) playerTestDB {
	t.Helper()

	//Create test DB
	tdb := testutils.CreateTestDB(t)

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
		"new1",
		"new2",
		"new3",
		"new4",
		"new5",
		"new6",
		"new7",
	}
	testutils.SetMockTwitchClient(validNames)

	//Players to add
	type playerStruct struct {
		DisplayName string
		TwitchName string
	}

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

	return playerTestDB{
		testDB: tdb,
		playerIDs: playerIDs,
	}
}

//Test AddPlayer
func TestAddPlayer(t *testing.T) {
	//Get test DB and handler
	tdb := initPlayerHandlerTestDB(t)
	h := &ReqHandler{tdb.testDB.Database}

	tests := playersGetPOSTTests()
	for _, test := range tests {
		testutils.CallMutationHandler(t, test, h.AddPlayer)
	}
}

//Test EditPlayer
func TestEditPlayer(t *testing.T) {
	//Get test DB and handler
	tdb := initPlayerHandlerTestDB(t)
	h := &ReqHandler{tdb.testDB.Database}

	tests := playersGetPATCHTests()
	for _, test := range tests {
		testutils.CallMutationHandler(t, test, h.EditPlayer)
	}
}

//Test GetPlayers
func TestGetPlayers(t *testing.T) {
	//Get teset DB and handler
	tdb := initPlayerHandlerTestDB(t)
	h := &ReqHandler{tdb.testDB.Database}

	tests := playersGetGETTests()

	for _, test := range tests {
		res := testutils.CallQueryHandler(t, test, h.GetPlayers)

		//If test failed, body doesn't have all this information
		resSuccess, ok := res["success"].(bool)
		if !ok {
			t.Fatalf("%s: unable to parse success field as boolean", test.TestName)
		}
		if !resSuccess {
			continue
		}

		//Confirm return type
		playerArr, ok := res["players"].([]any)
		if !ok {
			t.Errorf("%s: unable to parse races as array", test.TestName)
		}

		for _, a := range playerArr {
			obj, ok := a.(map[string]any)
			if !ok {
				t.Errorf("%s: unable to parse player objects", test.TestName)
				continue
			}

			//Validate object types
			playerName, ok := obj["name"].(string)
			if !ok {
				t.Errorf("%s: unable to parse player name as string", test.TestName)
				continue
			}
			twitchName, ok := obj["twitch_name"].(string)
			if !ok {
				t.Errorf("%s: unable to parse player twitch name as string", test.TestName)
				continue
			}

			//Make sure they match the params
			if len(test.URLParams["player_name"]) > 0 {
				validID := slices.Contains(test.URLParams["player_name"], playerName)
				if !validID {
					t.Errorf("%s: player name not filtered", test.TestName)
					continue
				}
			}

			if len(test.URLParams["twitch_name"]) > 0 {
				if !slices.Contains(test.URLParams["twitch_name"], twitchName) {
					t.Errorf("%s: twitch name not filtered", test.TestName)
					continue
				}
			}
		}
	}
}

//Helper to get POST tests
func playersGetPOSTTests() []testutils.MutationHandlerTest {
	//Create tests
	return []testutils.MutationHandlerTest{
		{
			TestName: "ValidAllFields",
			Body: map[string]any {
				"display_name": "newPlayer1",
				"twitch_name": "new1",
			},
			RequestType: "POST",
			Pattern: "POST /players",
			URL: "/players",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidEmptyDisplayName",
			Body: map[string]any {
				"display_name": "",
				"twitch_name": "new2",
			},
			RequestType: "POST",
			Pattern: "POST /players",
			URL: "/players",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidNoDisplayName",
			Body: map[string]any {
				"twitch_name": "new3",
			},
			RequestType: "POST",
			Pattern: "POST /players",
			URL: "/players",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "InvalidTwitchName",
			Body: map[string]any {
				"twitch_name": "invalid1",
			},
			RequestType: "POST",
			Pattern: "POST /players",
			URL: "/players",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidNoTwitchName",
			Body: map[string]any {
				"display_name": "NewPlayer2",
			},
			RequestType: "POST",
			Pattern: "POST /players",
			URL: "/players",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidEmptyTwitchName",
			Body: map[string]any {
				"display_name": "exprel",
				"twitch_name": "",
			},
			RequestType: "POST",
			Pattern: "POST /players",
			URL: "/players",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidTwitchNameInUse",
			Body: map[string]any {
				"display_name": "idk",
				"twitch_name": "expreli_",
			},
			RequestType: "POST",
			Pattern: "POST /players",
			URL: "/players",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidDisplayNameInUse",
			Body: map[string]any {
				"display_name": "Odme",
				"twitch_name": "new4",
			},
			RequestType: "POST",
			Pattern: "POST /players",
			URL: "/players",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidEmptyBody",
			Body: make(map[string]any),
			RequestType: "POST",
			Pattern: "POST /players",
			URL: "/players",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		},
	}
}

//Helper to get PATCH tests
func playersGetPATCHTests() []testutils.MutationHandlerTest {
	//Create tests
	return []testutils.MutationHandlerTest{
		{
			TestName: "ValidChangeName",
			Body: map[string]any {
				"display_name": "Odme2",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /players/{player_name}",
			URL: "/players/Odme",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidChangeTwitch",
			Body: map[string]any {
				"twitch_name": "new1",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /players/{player_name}",
			URL: "/players/jake",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidChangeBoth",
			Body: map[string]any {
				"display_name": "galax2",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /players/{player_name}",
			URL: "/players/galax",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidChangeNameFromTwitch",
			Body: map[string]any {
				"display_name": "fizz2",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /players/{player_name}",
			URL: "/players/fizz64",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "InvalidNameInUse",
			Body: map[string]any {
				"display_name": "me",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /players/{player_name}",
			URL: "/players/juka",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidNewTwitch",
			Body: map[string]any {
				"twitch_name": "invalid1",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /players/{player_name}",
			URL: "/players/jukatox",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidPathName",
			Body: map[string]any {
				"display_name": "test",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /players/{player_name}",
			URL: "/players/notreal",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		},
	}
}

func playersGetGETTests() []testutils.QueryHandlerTest{
	return []testutils.QueryHandlerTest{
		{
			TestName: "ValidAllPlayers",
			URLParams: make(map[string][]string),
			Pattern: "GET /players",
			URL: "/players",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidExpreliOdmeGamut",
			URLParams: map[string][]string {
				"player_name": {"expreli_", "Odme", "zgamut"},
			},
			Pattern: "GET /players",
			URL: "/players",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidSearchByTwitchNames",
			URLParams: map[string][]string {
				"twitch_name": {"expreli_", "clairdss", "jukatox"},
			},
			Pattern: "GET /players",
			URL: "/players",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			//Note that currently this handler tries to match BOTH. So me/clairdss is the only reponse here. It's kinda bad and should be fixed in general.
			TestName: "ValidSearchByBoth",
			URLParams: map[string][]string {
				"player_name": {"me"},
				"twitch_name": {"expreli_", "clairdss", "jukatox"},
			},
			Pattern: "GET /players",
			URL: "/players",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, 
	}
}