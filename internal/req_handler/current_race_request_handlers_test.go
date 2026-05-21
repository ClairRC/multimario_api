package req_handler

import (
	"net/http"
	"net/url"
	"testing"

	testutils "github.com/multimario_api/internal/testing"
	_ "github.com/ncruces/go-sqlite3/driver"
)

/*
* Tests for currentrace request handlers
 */

//Test SetPlayerCollectibleCount
func TestSetPlayerCollectibleCount(t *testing.T) {
	//Get test DB and handler
	tdb := initRecordHandlerTestDB(t)
	h := &ReqHandler{tdb.testDB.Database}

	tests := currentraceCollectibleGetPATCHTests()
	for _, test := range tests {
		testutils.CallMutationHandler(t, test, h.SetPlayerCollectibleCount)
	}
}

//Test UpdatePlayerGameTime
func TestUpdatePlayerGameTime(t *testing.T) {
	//Get test DB and handler
	tdb := initRecordHandlerTestDB(t)
	h := &ReqHandler{tdb.testDB.Database}

	tests := currentraceGameTimeGetPATCHTests()
	for _, test := range tests {
		testutils.CallMutationHandler(t, test, h.UpdatePlayerGameTime)
	}
}

//Test GetCurrentRaceStandings
func TestGetCurrentRaceStandings(t *testing.T) {
	//Get teset DB and handler
	tdb := initRecordHandlerTestDB(t)
	h := &ReqHandler{tdb.testDB.Database}

	tests := currentRaceStandingsGetGETTests()

	for _, test := range tests {
		res := testutils.CallQueryHandler(t, test, h.GetCurrentRaceStandings)

		//If test failed, body doesn't have all this information
		resSuccess, ok := res["success"].(bool)
		if !ok {
			t.Fatalf("%s: unable to parse success field as boolean", test.TestName)
		}
		if !resSuccess {
			continue
		}

		//Confirm return type
		playerArr, ok := res["standings"].([]any)
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
			_, ok = obj["num_collected"].(float64)
			if !ok {
				t.Errorf("%s: unable to parse num_collected as int", test.TestName)
				continue
			}

			//TODO: Test the rest of the parameters
		}
	}
}

//Test GetCurrentRaceStandings
func TestGetCurrentRaceRuns(t *testing.T) {
	//Get teset DB and handler
	tdb := initRecordHandlerTestDB(t)
	h := &ReqHandler{tdb.testDB.Database}

	tests := currentRaceRunsGetGETTests()

	for _, test := range tests {
		res := testutils.CallQueryHandler(t, test, h.GetCurrentRaceRuns)

		//If test failed, body doesn't have all this information
		resSuccess, ok := res["success"].(bool)
		if !ok {
			t.Fatalf("%s: unable to parse success field as boolean", test.TestName)
		}
		if !resSuccess {
			continue
		}

		//Confirm return type
		playerArr, ok := res["runs"].([]any)
		if !ok {
			t.Errorf("%s: unable to parse runs as array", test.TestName)
		}

		for _, a := range playerArr {
			obj, ok := a.(map[string]any)
			if !ok {
				t.Errorf("%s: unable to parse runs objects", test.TestName)
				continue
			}

			//Validate object types
			_, ok = obj["player_name"].(string)
			if !ok {
				t.Errorf("%s: unable to parse player name as string", test.TestName)
				continue
			}
			_, ok = obj["category_name"].(string)
			if !ok {
				t.Errorf("%s: unable to parse category as string", test.TestName)
				continue
			}
			_, ok = obj["estimate"].(string)
			if !ok {
				t.Errorf("%s: unable to parse estimate as string", test.TestName)
				continue
			}
			_, ok = obj["time"].(string)
			if !ok {
				t.Errorf("%s: unable to parse finish time as string", test.TestName)
				continue
			}
		}
	}
}

//Get Record Post Tests
func currentraceCollectibleGetPATCHTests() []testutils.MutationHandlerTest {
	return []testutils.MutationHandlerTest{
		{
			TestName: "Valid1",
			Body: map[string]any {
				"num_collected": 132,
			},
			RequestType: "PATCH",
			Pattern: "PATCH /currentrace/{player_name}",
			URL: "/currentrace/odme_",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "Valid2",
			Body: map[string]any {
				"num_collected": 541,
			},
			RequestType: "PATCH",
			Pattern: "PATCH /currentrace/{player_name}",
			URL: "/currentrace/expreli_",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "Valid3",
			Body: map[string]any {
				"num_collected": 539,
			},
			RequestType: "PATCH",
			Pattern: "PATCH /currentrace/{player_name}",
			URL: "/currentrace/thorn_97",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "InvalidPlayerDNE",
			Body: map[string]any {
				"num_collected": 539,
			},
			RequestType: "PATCH",
			Pattern: "PATCH /currentrace/{player_name}",
			URL: "/currentrace/invalid",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidPlayerNotInRace",
			Body: map[string]any {
				"num_collected": 539,
			},
			RequestType: "PATCH",
			Pattern: "PATCH /currentrace/{player_name}",
			URL: "/currentrace/galax",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidDatatype",
			Body: map[string]any {
				"num_collected": "hi",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /currentrace/{player_name}",
			URL: "/currentrace/expreli_",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidNoNumber",
			Body: map[string]any {
				"invalid": 500,
			},
			RequestType: "PATCH",
			Pattern: "PATCH /currentrace/{player_name}",
			URL: "/currentrace/expreli_",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, 
	}
}

func currentraceGameTimeGetPATCHTests() []testutils.MutationHandlerTest {
	return []testutils.MutationHandlerTest{
		{
			TestName: "Valid1",
			Body: map[string]any {
				"time": "0:30:12",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /currentrace/{player_name}/{category_name}",
			URL: "/currentrace/odme_/sm64_70",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "Valid2",
			Body: map[string]any {
				"time": "2:30:12",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /currentrace/{player_name}/{category_name}",
			URL: "/currentrace/thorn_97/sm3dw_"+url.PathEscape("any%"),
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "InvalidPlayerDNE",
			Body: map[string]any {
				"time": "0:15:12",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /currentrace/{player_name}/{category_name}",
			URL: "/currentrace/invalid/sm64_70",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidPlayerNotInRace",
			Body: map[string]any {
				"time": "0:15:12",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /currentrace/{player_name}/{category_name}",
			URL: "/currentrace/galax/sm64_70",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidTimeFormat",
			Body: map[string]any {
				"time": "15hr",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /currentrace/{player_name}/{category_name}",
			URL: "/currentrace/odme_/sm64_70",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidNoTime",
			Body: map[string]any {
				"not_time": "12:12:12",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /currentrace/{player_name}/{category_name}",
			URL: "/currentrace/odme_/sm64_70",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidCategoryNotInRace",
			Body: map[string]any {
				"time": "15:15:15",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /currentrace/{player_name}/{category_name}",
			URL: "/currentrace/odme_/sm64_120",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, 
	}
}

func currentRaceStandingsGetGETTests() []testutils.QueryHandlerTest {
	return []testutils.QueryHandlerTest{
		{
			TestName: "ValidAllStandings",
			URLParams: make(map[string][]string),
			Pattern: "GET /currentrace",
			URL: "/currentrace",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true, 
		}, {
			TestName: "ValidOnlyOdme",
			URLParams: map[string][]string{
				"player_name": {"Odme"},
			},
			Pattern: "GET /currentrace",
			URL: "/currentrace",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true, 
		}, 
	}
}

func currentRaceRunsGetGETTests() []testutils.QueryHandlerTest {
	return []testutils.QueryHandlerTest{
		{
			TestName: "ValidAllRuns",
			URLParams: make(map[string][]string),
			Pattern: "GET /records/runs",
			URL: "/records/runs",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true, 
		}, {
			TestName: "ValidSMORuns",
			URLParams: map[string][]string{
				"game_category": {"smo_any%"},
			},
			Pattern: "GET /records/runs",
			URL: "/records/runs",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true, 
		}, {
			TestName: "ValidSMG1And2Runs",
			URLParams: map[string][]string{
				"game_category": {"smg1_any%", "smg2_any%"},
			},
			Pattern: "GET /records/runs",
			URL: "/records/runs",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true, 
		}, {
			TestName: "ValidOdmeAndFizzRuns",
			URLParams: map[string][]string{
				"player_name": {"odme_", "fizz64"},
			},
			Pattern: "GET /records/runs",
			URL: "/records/runs",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true, 
		}, 
	}
}