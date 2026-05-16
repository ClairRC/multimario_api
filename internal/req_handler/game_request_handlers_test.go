package req_handler

import (
	"net/http"
	"slices"
	"testing"

	testutils "github.com/multimario_api/internal/testing"
)

//File for testing the Games request handler

//Test POST request
func TestAddGame(t *testing.T) {
	//Get handler and DB
	tbd := testutils.CreateTestDB(t)
	h := &ReqHandler{
		DataBase: tbd.Database,
	}

	tests := gamesGetPOSTTests()

	for _, test := range tests {
		//No need for the response since it just contains the error and success
		testutils.CallMutationHandler(t, test, h.AddGame)
	}
}

//Test UpdateRace
func TestChangeGameName(t *testing.T) {
	//Get test DB and handler
	tdb := testutils.CreateTestDB(t)
	h := &ReqHandler{tdb.Database}

	tests := gamesGetPATCHTests()
	for _, test := range tests {
		//No need for response
		testutils.CallMutationHandler(t, test, h.ChangeGameName)
	}
}

//Test GET /games
func TestGetGames(t *testing.T) {
	//Get teset DB and handler
	tdb := testutils.CreateTestDB(t)
	h := &ReqHandler{tdb.Database}

	tests := gamesGetGETTests()

	for _, test := range tests {
		res := testutils.CallQueryHandler(t, test, h.GetGames)

		//If test failed, body doesn't have all this information
		resSuccess, ok := res["success"].(bool)
		if !ok {
			t.Fatalf("%s: unable to parse success field as boolean", test.TestName)
		}
		if !resSuccess {
			continue
		}

		//Confirm return type
		gamesArr, ok := res["games"].([]any)
		if !ok {
			t.Errorf("%s: unable to parse games as array", test.TestName)
		}

		for _, a := range gamesArr {
			obj, ok := a.(map[string]any)
			if !ok {
				t.Errorf("%s: unable to parse game objects", test.TestName)
				continue
			}

			//Validate object types
			gameName, ok := obj["name"].(string)
			if !ok {
				t.Errorf("%s: unable to parse game name as string", test.TestName)
			}

			//Make sure they match the params
			if len(test.URLParams["name"]) > 0 {
				validName := slices.Contains(test.URLParams["name"], gameName)
				if !validName {
					t.Errorf("%s: game name not filtered", test.TestName)
					continue
				}
			}
		}
	}
}

func gamesGetPOSTTests() []testutils.MutationHandlerTest {
	return []testutils.MutationHandlerTest{
		{
			//Valid tests
			TestName: "ValidName",
			Body: map[string]any {
				"name": "oot",
			},
			RequestType: "POST",
			Pattern: "POST /games",
			URL: "/games",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "InvalidType",
			Body: map[string]any {
				"name": 1,
			},
			RequestType: "POST",
			Pattern: "POST /games",
			URL: "/games",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidNoName",
			Body: make(map[string]any),
			RequestType: "POST",
			Pattern: "POST /games",
			URL: "/games",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidWrongField",
			Body: map[string]any {
				"notName": "idk",
			},
			RequestType: "POST",
			Pattern: "POST /games",
			URL: "/games",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		},
	}
}

//Get patch tests for games
func gamesGetPATCHTests() []testutils.MutationHandlerTest {
	return []testutils.MutationHandlerTest{
		{
			TestName: "ValidChangeMario64",
			Body: map[string]any {
				"name": "sm74",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /games/{game_name}",
			URL: "/games/sm64",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidChangeSM3DW",
			Body: map[string]any {
				"name": "sm4dw",
			}, 
			RequestType: "PATCH",
			Pattern: "PATCH /games/{game_name}",
			URL: "/games/sm3dw",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "InvalidType",
			Body: map[string]any {
				"name": 1,
			},
			RequestType: "PATCH",
			Pattern: "PATCH /games/{game_name}",
			URL: "/games/sms",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidNoName",
			Body: make(map[string]any),
			RequestType: "PATCH",
			Pattern: "PATCH /games/{game_name}",
			URL: "/games/smo",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidBadPathValue",
			Body: map[string]any {
				"game": "star_road",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /games/{game_name}",
			URL: "/games/sm67",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidGameAlreadyExists",
			Body: map[string]any {
				"name": "smg2",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /games/{game_name}",
			URL: "/games/smg1",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		},
	}
}

//Get GET tests
func gamesGetGETTests() []testutils.QueryHandlerTest {
	return []testutils.QueryHandlerTest{
		{
			TestName: "ValidGetSM64",
			URLParams: map[string][]string {
				"name": {"sm64"},
			},
			Pattern: "GET /games",
			URL: "/games",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidGet3DW", 
			URLParams: map[string][]string {
				"name": {"sm3dw"},
			},
			Pattern: "GET /games",
			URL: "/games",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidGetSMSorSMG",
			URLParams: map[string][]string {
				"name": {"sms", "smg1"},
			},
			Pattern: "GET /games",
			URL: "/games",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidNoRestrictions",
			URLParams: make(map[string][]string),
			Pattern: "GET /games",
			URL: "/games",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, { 
			TestName: "ValidNoGameExists",
			URLParams: map[string][]string {
				"name": {"none"},
			},
			Pattern: "GET /games",
			URL: "/games",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		},
	}
}
