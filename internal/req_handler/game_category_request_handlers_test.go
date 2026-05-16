package req_handler

import (
	"net/http"
	"net/url"
	"slices"
	"testing"

	testutils "github.com/multimario_api/internal/testing"
)

//Tests for game category handlers

func TestAddGameCategory(t *testing.T) {
	//Get handler and DB
	tbd := testutils.CreateTestDB(t)
	h := &ReqHandler{
		DataBase: tbd.Database,
	}

	tests := gameCategoriesGetPOSTTests()

	for _, test := range tests {
		//No need for the response since it just contains the error and success
		testutils.CallMutationHandler(t, test, h.AddGameCategory)
	}
}

//Test UpdateGameCategory
func TestEditGameCategory(t *testing.T) {
	//Get test DB and handler
	tdb := testutils.CreateTestDB(t)
	h := &ReqHandler{tdb.Database}

	tests := gameCategoriesGetPATCHTests()
	for _, test := range tests {
		//No need for response
		testutils.CallMutationHandler(t, test, h.EditGameCategory)
	}
}

//Test get game categories
func TestGetGameCategories(t *testing.T) {
	//Get teset DB and handler
	tdb := testutils.CreateTestDB(t)
	h := &ReqHandler{tdb.Database}

	tests := gameCategoriesGetGETTests()

	for _, test := range tests {
		res := testutils.CallQueryHandler(t, test, h.GetGameCategories)

		//If test failed, body doesn't have all this information
		resSuccess, ok := res["success"].(bool)
		if !ok {
			t.Fatalf("%s: unable to parse success field as boolean", test.TestName)
		}
		if !resSuccess {
			continue
		}

		//Confirm return type
		gameCatArr, ok := res["game_categories"].([]any)
		if !ok {
			t.Errorf("%s: unable to parse games as array", test.TestName)
		}

		for _, a := range gameCatArr {
			obj, ok := a.(map[string]any)
			if !ok {
				t.Errorf("%s: unable to parse game category objects", test.TestName)
				continue
			}

			//Validate object types
			gameCatName, ok := obj["name"].(string)
			if !ok {
				t.Errorf("%s: unable to parse game category name as string", test.TestName)
			}

			_, ok = obj["id"].(float64)
			if !ok {
				t.Errorf("%s: unable to parse game category id as int", test.TestName)
			}

			gameName, ok := obj["game"].(string)
			if !ok {
				t.Errorf("%s: unable to parse game name as string", test.TestName)
			}

			//Estimate is allowed to be nil
			if obj["estimate"] != nil {
				_, ok = obj["estimate"].(string)
				if !ok {
					t.Errorf("%s: unable to parse estimate as string", test.TestName)
				}
			}

			_, ok = obj["num_collectibles"].(float64)
			if !ok {
				t.Errorf("%s: unable to parse num_collectbiles as int", test.TestName)
			}

			//Make sure they match the params
			//TODO: Add logic to test that the race categories are correct. That would be a lot of extra work
			//so for now im just gonna confirm that with my own two eyes, but adding that logic to these tests would be nice
			if len(test.URLParams["name"]) > 0 {
				validName := slices.Contains(test.URLParams["name"], gameCatName)
				if !validName {
					t.Errorf("%s: game category name not filtered", test.TestName)
					continue
				}
			}

			if len(test.URLParams["game"]) > 0 {
				validGame := slices.Contains(test.URLParams["game"], gameName)
				if !validGame {
					t.Errorf("%s: game name not filtered", test.TestName)
				}
			}
		}
	}
}

func gameCategoriesGetPOSTTests() []testutils.MutationHandlerTest {
	return []testutils.MutationHandlerTest{
		{
			TestName: "Valid",
			Body: map[string]any {
				"category_name": "smg2_greens",
				"game_name": "smg2",
				"estimate": "3:30:00",
				"num_collectibles": 120,
			},
			RequestType: "POST",
			Pattern: "POST /gamecategories",
			URL: "/gamecategories",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "InvalidNameType",
			Body: map[string]any {
				"category_name": 3,
				"game_name": "sm64",
				"estimate": "00:55:00",
				"num_collectibles": 90, 
			},
			RequestType: "POST",
			Pattern: "POST /gamecategories",
			URL: "/gamecategories",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidCollectiblesType",
			Body: map[string]any {
				"category_name": "oot_100%",
				"game_name": "sms",
				"estimate": "10:00:00",
				"num_collectibles": "100",
			},
			RequestType: "POST",
			Pattern: "POST /gamecategories",
			URL: "/gamecategories",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidEstimateFormat",
			Body: map[string]any {
				"category_name": "smo_darkest_side",
				"game_name": "smo",
				"estimate": "9 hours",
				"num_collectibles": 753,
			},
			RequestType: "POST",
			Pattern: "POST /gamecategories",
			URL: "/gamecategories",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidEstimateType",
			Body: map[string]any {
				"category_name": "smo_darkest_side",
				"game_name": "smo",
				"estimate": 9,
				"num_collectibles": 753,
			},
			RequestType: "POST",
			Pattern: "POST /gamecategories",
			URL: "/gamecategories",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidGameDoesntExist",
			Body: map[string]any {
				"category_name": "cat",
				"game_name": "mm",
				"estimate": "55:55:55",
				"num_collectibles": 12,
			}, 
			RequestType: "POST",
			Pattern: "POST /gamecategories",
			URL: "/gamecategories",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
			}, {
				TestName: "InvalidMissingGame",
				Body: map[string]any {
					"category_name": "360",
					"estimate": "1:00:00",
					"num_collectibles": 1431,
				},
			RequestType: "POST",
			Pattern: "POST /gamecategories",
			URL: "/gamecategories",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		},
	}
}

//Get PATCh tests
func gameCategoriesGetPATCHTests() []testutils.MutationHandlerTest {
	return []testutils.MutationHandlerTest{
		{
			TestName: "ValidAllValues",
			Body: map[string]any {
				"category_name": "sm64_any%",
				"game_name": "sm64",
				"num_collectibles": 70,
				"estimate": "1:00:00",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /gamecategories/{game_category_name}",
			URL: "/gamecategories/sm64_70",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidNoNewName",
			Body: map[string]any {
				"game_name": "smg1",
				"num_collectibles": 79,
				"estimate": "1:00:00",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /gamecategories/{game_category_name}",
			URL: "/gamecategories/sm64_120",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidSameGame",
			Body: map[string]any {
				"category_name": "smg1_luigi_time",
				"num_collectibles": 61,
				"estimate": "13:50:00",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /gamecategories/{game_category_name}",
			URL: "/gamecategories/" + url.PathEscape("smg1_any%"), //Need to escape the %
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidSameCollectbiles",
			Body: map[string]any {
				"category_name": "smg1_50%",
				"estimate": "9:00:00",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /gamecategories/{game_category_name}",
			URL: "/gamecategories/smg1_120",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidNoChanges",
			Body: make(map[string]any),
			RequestType: "PATCH",
			Pattern: "PATCH /gamecategories/{game_category_name}",
			URL: "/gamecategories/" + url.PathEscape("sms_any%"),
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidNewEstimateNewCollectibles",
			Body: map[string]any {
				"num_collectibles": 119,
				"estimate": "2:00:00",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /gamecategories/{game_category_name}",
			URL: "/gamecategories/sms_120",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "InvalidCategoryDoesntExist",
			Body: map[string]any {
				"category_name": "sm64_any%",
				"game_name": "sm64",
				"num_collectibles": 70,
				"estimate": "1:00:00",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /gamecategories/{game_category_name}",
			URL: "/gamecategories/sm64_71",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidCategoryAlreadyExists",
			Body: map[string]any {
				"category_name": "sm3dw_100%",
				"game_name": "sm3dw",
				"num_collectibles": 380,
				"estimate": "1:00:00",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /gamecategories/{game_category_name}",
			URL: "/gamecategories/sm3dw_380",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidGameDoesNotExist",
			Body: map[string]any {
				"category_name": "sm64_2%",
				"game_name": "oot",
				"num_collectibles": 423,
				"estimate": "1:10:00",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /gamecategories/{game_category_name}",
			URL: "/gamecategories/" + url.PathEscape("smo_any%"),
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidType",
			Body: map[string]any {
				"category_name": 12,
			},
			RequestType: "PATCH",
			Pattern: "PATCH /gamecategories/{game_category_name}",
			URL: "/gamecategories/smo_all_moons",
						ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, 
	}
}

func gameCategoriesGetGETTests() []testutils.QueryHandlerTest {
	return []testutils.QueryHandlerTest{
		{ 
			TestName: "ValidAllCategories",
			URLParams: make(map[string][]string),
			Pattern: "GET /gamecategories",
			URL: "/gamecategories",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidSM64Cats",
			URLParams: map[string][]string {
				"game": {"sm64"},
			},
			Pattern: "GET /gamecategories",
			URL: "/gamecategories",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidSMG2Cats",
			URLParams: map[string][]string {
				"game": {"smg2"},
			},
			Pattern: "GET /gamecategories",
			URL: "/gamecategories",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidSMG2OrSMOCats",
			URLParams: map[string][]string {
				"game": {"smg2", "smo"},
			},
			Pattern: "GET /gamecategories",
			URL: "/gamecategories",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "Valid602Cats",
			URLParams: map[string][]string {
				"race_category": {"602"},
			},
			Pattern: "GET /gamecategories",
			URL: "/gamecategories",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidSMG1_120",
			URLParams: map[string][]string {
				"name": {"smg1_120"},
			},
			Pattern: "GET /gamecategories",
			URL: "/gamecategories",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidSMOCatsIn602",
			URLParams: map[string][]string {
				"game": {"smo"},
				"race_category": {"602"},
			},
			Pattern: "GET /gamecategories",
			URL: "/gamecategories",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidSMGCatsIn246/540/602",
			URLParams: map[string][]string {
				"game": {"smg2", "smg1"},
				"race_category": {"246", "540", "602"},
			},
			Pattern: "GET /gamecategories",
			URL: "/gamecategories",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidSMOCatsIn540/sandbox_100%",
			URLParams: map[string][]string {
				"game": {"smo"},
				"race_category": {"540", "sandbox_100%"},
			},
			Pattern: "GET /gamecategories",
			URL: "/gamecategories",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidSMOany%SMG2120",
			URLParams: map[string][]string {
				"name": {"smo_any%", "smg2_120"},
			},
			Pattern: "GET /gamecategories",
			URL: "/gamecategories",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidCategoriesIn1862",
			URLParams: map[string][]string {
				"race_category": {"1862"},
			},
			Pattern: "GET /gamecategories",
			URL: "/gamecategories",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		},
	}
}