package req_handler

import (
	"net/http"
	"net/url"
	"testing"

	testutils "github.com/multimario_api/internal/testing"
	_ "github.com/ncruces/go-sqlite3/driver"
)

/*
* Tests for record runs handlers
 */

//Test EditRun
func TestEditRun(t *testing.T) {
	//Get test DB and handler

	//This uses the same DB as record request tests because runs are just a subset of records
	tdb := initRecordHandlerTestDB(t)
	h := &ReqHandler{tdb.testDB.Database}

	tests := runsGetPATCHTests()
	for _, test := range tests {
		testutils.CallMutationHandler(t, test, h.EditRun)
	}
}

//Test GetRuns
func TestGetRuns(t *testing.T) {
	//Get teset DB and handler
	tdb := initRecordHandlerTestDB(t)
	h := &ReqHandler{tdb.testDB.Database}

	tests := runsGetGETTests()

	for _, test := range tests {
		res := testutils.CallQueryHandler(t, test, h.GetRuns)

		//If test failed, body doesn't have all this information
		resSuccess, ok := res["success"].(bool)
		if !ok {
			t.Fatalf("%s: unable to parse success field as boolean", test.TestName)
		}
		if !resSuccess {
			continue
		}

		//Validate parameters
		resRuns, ok := res["runs"].([]any)
		if !ok {
			t.Errorf("%s: unable to parse runs as array", test.TestName)
			continue
		}

		for _, r := range resRuns {
			run, ok := r.(map[string]any)
			if !ok {
				t.Errorf("%s: unable to parse run as json object", test.TestName)
				continue
			}

			_, ok = run["id"].(float64)
			if !ok {
				t.Errorf("%s: unable to parse run id as int", test.TestName)
				continue
			}

			_, ok = run["player_name"].(string)
			if !ok {
				t.Errorf("%s: unable to parse run player name as string", test.TestName)
				continue
			}

			_, ok = run["race_id"].(float64)
			if !ok {
				t.Errorf("%s: unable to parse run race id as int", test.TestName)
				continue
			}

			_, ok = run["game_category"].(string)
			if !ok {
				t.Errorf("%s: unable to parse game category as string", test.TestName)
				continue
			}

			_, ok = run["estimate"].(string)
			if !ok {
				t.Errorf("%s: unable to parse run estimate as string", test.TestName)
				continue
			}
			
			_, ok = run["time"].(string)
			if !ok {
				t.Errorf("%s: unable to parse run time as string", test.TestName)
				continue
			}
		} 
	}
}

//Get Record Post Tests
func runsGetPATCHTests() []testutils.MutationHandlerTest {
	return []testutils.MutationHandlerTest {
		{
			TestName: "ValidAllFields",
			Body: map[string]any {
				"time": "1:59:59",
				"estimate": "1:49:59",
				"run_num": 2,
			},
			RequestType: "PATCH",
			Pattern: "PATCH /records/{race_id}/{player_name}/runs/{category_name}",
			URL: "/records/1/odme_/runs/sm64_120",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidJustTime",
			Body: map[string]any {
				"time": "5:59:59",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /records/{race_id}/{player_name}/runs/{category_name}",
			URL: "/records/2/odme_/runs/smg2_" + url.QueryEscape("any%"),
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidJustEstimate",
			Body: map[string]any {
				"estimate": "0:40:59",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /records/{race_id}/{player_name}/runs/{category_name}",
			URL: "/records/3/odme_/runs/sm64_70",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "InvalidRace",
			Body: map[string]any {
				"estimate": "0:40:59",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /records/{race_id}/{player_name}/runs/{category_name}",
			URL: "/records/99/clairdss/runs/sm64_70",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidPlayer",
			Body: map[string]any {
				"estimate": "0:40:59",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /records/{race_id}/{player_name}/runs/{category_name}",
			URL: "/records/1/notreal/runs/smo_all_moons",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidRecord",
			Body: map[string]any {
				"time": "0:40:59",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /records/{race_id}/{player_name}/runs/{category_name}",
			URL: "/records/1/jukatox/runs/sm64_120",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidCategory",
			Body: map[string]any {
				"estimate": "0:40:59",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /records/{race_id}/{player_name}/runs/{category_name}",
			URL: "/records/1/Odme/runs/smo_880",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidCategoryNotInRecord",
			Body: map[string]any {
				"estimate": "0:40:59",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /records/{race_id}/{player_name}/runs/{category_name}",
			URL: "/records/1/Odme/runs/sm64_70",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidCategoryNotInRecord2",
			Body: map[string]any {
				"time": "1:40:59",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /records/{race_id}/{player_name}/runs/{category_name}",
			URL: "/records/2/galax_v/runs/smo_"+url.QueryEscape("any%"),
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidTimeFormat",
			Body: map[string]any {
				"time": "1:40",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /records/{race_id}/{player_name}/runs/{category_name}",
			URL: "/records/2/galax_v/runs/sms_"+url.QueryEscape("any%"),
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidEstimateFormat",
			Body: map[string]any {
				"estimate": "1:40",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /records/{race_id}/{player_name}/runs/{category_name}",
			URL: "/records/2/galax_v/runs/sms_"+url.QueryEscape("any%"),
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, 
	}
}

func runsGetGETTests() []testutils.QueryHandlerTest {
	return []testutils.QueryHandlerTest{
		{
			TestName: "ValidAllRuns",
			URLParams: make(map[string][]string),
			Pattern: "GET /records/runs",
			URL: "/records/runs",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true, 
		}, {
			TestName: "ValidAllRunsP2",
			URLParams: map[string][]string{
				"page_num": {"2"},
			},
			Pattern: "GET /records/runs",
			URL: "/records/runs",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true, 
		}, {
			TestName: "ValidRaceID12",
			URLParams: map[string][]string{
				"race_id": {"12"},
			},
			Pattern: "GET /records/runs",
			URL: "/records/runs",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true, 
		}, {
			TestName: "ValidRaceID2",
			URLParams: map[string][]string{
				"race_id": {"2"},
			},
			Pattern: "GET /records/runs",
			URL: "/records/runs",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true, 
		}, {
			TestName: "ValidRaceID1",
			URLParams: map[string][]string{
				"race_id": {"1"},
			},
			Pattern: "GET /records/runs",
			URL: "/records/runs",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true, 
		}, {
			TestName: "ValidRaceID1Or2",
			URLParams: map[string][]string{
				"race_id": {"1", "2"},
			},
			Pattern: "GET /records/runs",
			URL: "/records/runs",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true, 
		}, {
			TestName: "ValidRaceSM64120",
			URLParams: map[string][]string{
				"game_category": {"sm64_120"},
			},
			Pattern: "GET /records/runs",
			URL: "/records/runs",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true, 
		}, {
			TestName: "ValidRaceSM64120Or70",
			URLParams: map[string][]string{
				"game_category": {"sm64_120", "sm64_70"},
			},
			Pattern: "GET /records/runs",
			URL: "/records/runs",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true, 
		}, {
			TestName: "ValidFromOdme",
			URLParams: map[string][]string{
				"player_name": {"Odme"},
			},
			Pattern: "GET /records/runs",
			URL: "/records/runs",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true, 
		}, {
			TestName: "ValidGalaxV",
			URLParams: map[string][]string{
				"player_name": {"galax_v"},
			},
			Pattern: "GET /records/runs",
			URL: "/records/runs",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true, 
		}, {
			TestName: "ValidJukaOrThorn",
			URLParams: map[string][]string{
				"player_name": {"jukatox", "jake"},
			},
			Pattern: "GET /records/runs",
			URL: "/records/runs",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true, 
		}, {
			TestName: "ValidFieldsDontExist",
			URLParams: map[string][]string{
				"player_name": {"invalid"},
				"game_category": {"invalid"},
				"race_id": {"999"},
			},
			Pattern: "GET /records/runs",
			URL: "/records/runs",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true, 
		}, {
			TestName: "ValidSM64SMG1120RunsFromOdmeOrClairDSS",
			URLParams: map[string][]string{
				"player_name": {"odme_", "me"},
				"game_category": {"sm64_120", "smg1_120"},
			},
			Pattern: "GET /records/runs",
			URL: "/records/runs",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true, 
		}, 
	}
}