package req_handler

import (
	"net/http"
	"net/url"
	"slices"
	"testing"

	testutils "github.com/multimario_api/internal/testing"
)

//Tests for race category handlers

func TestAddRaceCategory(t *testing.T) {
	//Get handler and DB
	tbd := testutils.CreateTestDB(t)
	h := &ReqHandler{
		DataBase: tbd.Database,
	}

	tests := raceCategoriesGetPOSTTests()

	for _, test := range tests {
		//No need for the response since it just contains the error and success
		testutils.CallMutationHandler(t, test, h.AddRaceCategory)
	}
}

//Test UpdateRaceCategory
func TestEditRaceCategory(t *testing.T) {
	//Get test DB and handler
	tdb := testutils.CreateTestDB(t)
	h := &ReqHandler{tdb.Database}

	tests := raceCategoriesGetPATCHTests()
	for _, test := range tests {
		//No need for response
		testutils.CallMutationHandler(t, test, h.EditRaceCategory)
	}
}

//Test get game categories
func TestGetRaceCategories(t *testing.T) {
	//Long and dumb function. Sorry.

	//Get test DB and handler
	tdb := testutils.CreateTestDB(t)
	h := &ReqHandler{tdb.Database}

	tests := raceCategoriesGetGETTests()

	for _, test := range tests {
		res := testutils.CallQueryHandler(t, test, h.GetRaceCategories)

		//If test failed, body doesn't have all this information
		resSuccess, ok := res["success"].(bool)
		if !ok {
			t.Fatalf("%s: unable to parse success field as boolean", test.TestName)
		}
		if !resSuccess {
			continue
		}

		//Confirm return type
		raceCatArr, ok := res["race_categories"].([]any)
		if !ok {
			t.Errorf("%s: unable to parse race categories as array", test.TestName)
		}
		
		for _, a := range raceCatArr {
			obj, ok := a.(map[string]any)
			if !ok {
				t.Errorf("%s: unable to parse race category objects", test.TestName)
				continue
			}

			//Validate object types
			raceCatName, ok := obj["name"].(string)
			if !ok {
				t.Errorf("%s: unable to parse race category name as string", test.TestName)
			}

			_, ok = obj["num_collectibles"].(float64)
			if !ok {
				t.Errorf("%s: unable to parse num collectibles as int", test.TestName)
			}

			//Check inner game category array
			//TODO: This test doesn't actually check if the game/categories were filtered. That could be changed.
			_, ok = obj["game_categories"].([]any)
			if !ok {
				t.Errorf("%s: unable to parse game categories as array", test.TestName)
			}

			//Make sure they match the params
			if len(test.URLParams["race_category"]) > 0 {
				validName := slices.Contains(test.URLParams["race_category"], raceCatName)
				if !validName {
					t.Errorf("%s: game category name not filtered", test.TestName)
					continue
				}
			}
		}
	}
}

func raceCategoriesGetPOSTTests() []testutils.MutationHandlerTest {
	return []testutils.MutationHandlerTest{
		{
			TestName: "Valid",
			Body: map[string]any {
				"name": "test1",
				"game_categories": []any {
					map[string]any {
						"game_category_name": "smg2_120",
					},
					map[string]any {
						"game_category_name": "smg1_120",
					},
				},
			},
			RequestType: "POST",
			Pattern: "POST /racecategories",
			URL: "/racecategories",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidOneGame",
			Body: map[string]any {
				"name": "test2",
				"game_categories": []any {
					map[string]any {
						"game_category_name": "smg2_120",
					},
				},
			},
			RequestType: "POST",
			Pattern: "POST /racecategories",
			URL: "/racecategories",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "InvalidNoGames",
			Body: map[string]any {
				"name": "test3",
				"game_categories": make([]any, 0),
			},
			RequestType: "POST",
			Pattern: "POST /racecategories",
			URL: "/racecategories",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidNoGameCatNames",
			Body: map[string]any {
				"name": "test4",
				"game_categories": []any {
					make(map[string]any),
				},
			},
			RequestType: "POST",
			Pattern: "POST /racecategories",
			URL: "/racecategories",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidGameCatDoesntExist",
			Body: map[string]any {
				"name": "test5",
				"game_categories": []any {
					map[string]any {
						"game_category_name": "smg3_120",
					},
					map[string]any {
						"game_category_name": "smg1_120",
					},
				},
			},
			RequestType: "POST",
			Pattern: "POST /racecategories",
			URL: "/racecategories",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidMissingRaceCategoryName",
			Body: map[string]any {
				"game_categories": []any {
					map[string]any {
						"game_category_name": "sms_any%",
					},
					map[string]any {
						"game_category_name": "sm64_70",
					},
				},
			},
			RequestType: "POST",
			Pattern: "POST /racecategories",
			URL: "/racecategories",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidRaceCatAlreadyExists",
			Body: map[string]any {
				"name": "602",
				"game_categories": []any {
					map[string]any {
						"game_category_name": "smg3_120",
					},
					map[string]any {
						"game_category_name": "smg1_120",
					},
				},
			},
			RequestType: "POST",
			Pattern: "POST /racecategories",
			URL: "/racecategories",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		},
	}
}

//Get PATCh tests
func raceCategoriesGetPATCHTests() []testutils.MutationHandlerTest {
	return []testutils.MutationHandlerTest{
		{
			TestName: "ValidChangeNameAndGames",
			Body: map[string]any {
				"name": "603",
				"game_categories": []any {
					map[string]any {
						"game_category_name": "smg2_120",
					},
					map[string]any {
						"game_category_name": "smo_100%",
					},
					map[string]any {
						"game_category_name": "sm3dw_100%",
					},
				},
			},
			RequestType: "PATCH",
			Pattern: "PATCH /racecategories/{race_category_name}",
			URL: "/racecategories/602",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidChangeNameNotGames",
			Body: map[string]any {
				"name": "247",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /racecategories/{race_category_name}",
			URL: "/racecategories/246",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidChangeGames",
			Body: map[string]any {
				"game_categories": []any {
					map[string]any {
						"game_category_name": "sm3dw_380",
					},
				},
			},
			RequestType: "PATCH",
			Pattern: "PATCH /racecategories/{race_category_name}",
			URL: "/racecategories/540",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidEmpty",
			Body: make(map[string]any),
			RequestType: "PATCH",
			Pattern: "PATCH /racecategories/{race_category_name}",
			URL: "/racecategories/540",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "ValidEmptyArray",
			Body: map[string]any {
				"name": "6111",
				"game_categories": make([]any, 0),
			},
			RequestType: "PATCH",
			Pattern: "PATCH /racecategories/{race_category_name}",
			URL: "/racecategories/1862",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, {
			TestName: "InvalidGameCatDoesntExist",
			Body: map[string]any {
				"game_categories": []any {
					map[string]any {
						"game_category_name": "sm3dw_notreal",
					},
				},
			},
			RequestType: "PATCH",
			Pattern: "PATCH /racecategories/{race_category_name}",
			URL: "/racecategories/" + url.PathEscape("sandbox_100%"),
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidRaceCatDoesntExist",
			Body: map[string]any {
				"game_categories": []any {
					map[string]any {
						"game_category_name": "sm3dw_notreal",
					},
				},
			},
			RequestType: "PATCH",
			Pattern: "PATCH /racecategories/{race_category_name}",
			URL: "/racecategories/sandbox_95",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, {
			TestName: "InvalidRaceCatAlreadyExists",
			Body: map[string]any {
				"name": "sandbox_any%",
			},
			RequestType: "PATCH",
			Pattern: "PATCH /racecategories/{race_category_name}",
			URL: "/racecategories/247",
			ExpectedResponseCode: http.StatusBadRequest,
			ExpectedSuccess: false,
		}, 
	}
}

func raceCategoriesGetGETTests() []testutils.QueryHandlerTest {
	return []testutils.QueryHandlerTest{
		{ 
			TestName: "ValidAll",
			URLParams: make(map[string][]string),
			Pattern: "GET /racecategories",
			URL: "/racecategories",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, { 
			TestName: "Valid602",
			URLParams: map[string][]string {
				"race_category": {"602"},
			},
			Pattern: "GET /racecategories",
			URL: "/racecategories",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, { 
			TestName: "Valid6021862",
			URLParams: map[string][]string {
				"race_category": {"602", "1862"},
			},
			Pattern: "GET /racecategories",
			URL: "/racecategories",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, { 
			TestName: "ValidRaceCategoriesWithSMO",
			URLParams: map[string][]string {
				"game": {"smo"},
			},
			Pattern: "GET /racecategories",
			URL: "/racecategories",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, { 
			TestName: "ValidCategoryWithSMSany%",
			URLParams: map[string][]string {
				"game_category": {"sms_any%"},
			},
			Pattern: "GET /racecategories",
			URL: "/racecategories",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, { 
			TestName: "ValidHasSMOAndSMOAllMoons",
			URLParams: map[string][]string {
				"game_category": {"smo_all_moons"},
				"game": {"smo"},
			},
			Pattern: "GET /racecategories",
			URL: "/racecategories",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, { 
			TestName: "Valid602AndSMO",
			URLParams: map[string][]string {
				"race_category": {"602"},
				"game": {"smo"},
			},
			Pattern: "GET /racecategories",
			URL: "/racecategories",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, { 
			TestName: "ValidSM6470OrSMG1Any%",
			URLParams: map[string][]string {
				"game_category": {"sm64_70", "smg1_any%"},
			},
			Pattern: "GET /racecategories",
			URL: "/racecategories",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, { 
			TestName: "ValidSMOAny%InSandboxAny%",
			URLParams: map[string][]string {
				"race_category": {"sandbox_any%"},
				"game_category": {"smo_any%"},
			},
			Pattern: "GET /racecategories",
			URL: "/racecategories",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		}, { 
			TestName: "ValidCategoriesWithSMG2",
			URLParams: map[string][]string {
				"game": {"smg2"},
			},
			Pattern: "GET /racecategories",
			URL: "/racecategories",
			ExpectedResponseCode: http.StatusOK,
			ExpectedSuccess: true,
		},
	}
}