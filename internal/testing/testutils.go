package testutils

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/multimario_api/internal/db"
)

//Package for general testing functionality

//Test structs for POST/PATCH handlers
type MutationHandlerTest struct {
	TestName string
	Body map[string]any
	RequestType string
	Pattern string
	URL string
	ExpectedResponseCode int
	ExpectedSuccess bool
}

//Test struct specifically for GET because they are different and weird
type QueryHandlerTest struct {
	TestName string
	URLParams map[string][]string
	Pattern string
	URL string
	ExpectedResponseCode int
	ExpectedSuccess bool
}

//Test database struct
type TestDB struct {
	Database *sql.DB
	GameIDs map[string]int64 
	GameCatIDs map[string]int64
	RaceCatIDs map[string]int64
}

//Creates and seeds database with a handful of games and categories that actually exist for testing.
func CreateTestDB(t *testing.T) TestDB {
	t.Helper()

	//Create database
	//Open database
	database, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("unable to initialize test database: %v", err)
	}
	err = db.DatabaseInit(database)
	if err != nil {
		t.Fatalf("unable to initialize test database: %v", err)
	}

	//Cleanup after the test
	t.Cleanup(func() {
		database.Close()
	})

	out := TestDB{
		Database: database,
		GameIDs: make(map[string]int64),
		GameCatIDs: make(map[string]int64),
		RaceCatIDs: make(map[string]int64),
	}

	//Seed db
	//Test db information
	games := []string {
		"sm64",
		"smg1",
		"sms",
		"smg2",
		"smo",
		"sm3dw",
	}

	type gameCategories struct {
		game string
		gameCat string
		estimate string
		numCollectibles int
	}
	gameCats := []gameCategories {
		{"sm64", "sm64_70", "1:00:00", 70},
		{"sm64", "sm64_120", "2:00:00", 120},
		{"smg1", "smg1_any%", "3:00:00", 61},
		{"smg1", "smg1_120", "6:00:00", 120},
		{"sms", "sms_any%", "1:30:00", 44},
		{"sms", "sms_120", "3:30:00", 120},
		{"smg2", "smg2_any%", "3:30:00", 71},
		{"smg2", "smg2_120", "6:00:00", 120},
		{"smg2", "smg2_242", "10:00:00", 242},
		{"smo", "smo_any%", "1:30:00", 124},
		{"smo", "smo_darker_side", "3:30:00", 503},
		{"smo", "smo_all_moons", "9:00:00", 880},
		{"smo", "smo_100%", "11:00:00", 880},
		{"sm3dw", "sm3dw_any%", "1:50:00", 170},
		{"sm3dw", "sm3dw_380", "4:00:00", 380},
		{"sm3dw", "sm3dw_100%", "8:00:00", 380},
	}

	type raceCategories struct {
		gameCats []string
		raceCat string
	}
	raceCats := []raceCategories {
		{[]string{"sm64_120", "smg1_120", "sms_120", "smg2_242"}, "602"},
		{[]string{"sm64_70", "smg1_any%", "sms_any%", "smg2_any%"}, "246"},
		{[]string{"sm64_70", "sms_any%", "smo_any%"}, "sandbox_any%"},
		{[]string{"sm64_120", "sms_120", "smo_all_moons"}, "sandbox_100%"},
		{[]string{"sm64_120", "smg1_120", "sms_120", "smg2_242", "sm3dw_380", "smo_all_moons"}, "1862"},
		{[]string{"sm64_70", "smg1_any%", "sms_any%", "smg2_any%", "sm3dw_any%", "smo_any%"}, "540"},
	}

	//Games
	for _, game := range games {
		stmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (?)", db.TableGames, db.ColGameName)

		res, err := database.Exec(stmt, game)
		if err != nil {
			t.Fatalf("unable to initialize test database: %v", err)
		}
		out.GameIDs[game], err = res.LastInsertId()
		if err != nil {
			t.Fatalf("unable to initialize test database: %v", err)
		}
	}

	//Game categories
	for _, gc := range gameCats {
		stmt := fmt.Sprintf("INSERT INTO %s (%s, %s, %s, %s) VALUES (?, ?, ?, ?)", 
			db.TableGameCategories, db.ColGameCategoryGameID, db.ColGameCategoryName, db.ColGameCategoryEstimate, db.ColGameCategoryNumCollectibles)
		args := []any{out.GameIDs[gc.game], gc.gameCat, gc.estimate, gc.numCollectibles}

		res, err := database.Exec(stmt, args...)
		if err != nil {
			t.Fatalf("unable to initialize test database: %v", err)
		}
		out.GameCatIDs[gc.gameCat], err = res.LastInsertId()
		if err != nil {
			t.Fatalf("unable to initialize test database: %v", err)
		}
	}

	//Race categories and linking table
	for _, rc := range raceCats {
		stmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (?)", db.TableRaceCategories, db.ColRaceCategoryName)
		res, err := database.Exec(stmt, rc.raceCat)

		if err != nil {
			t.Fatalf("unable to initialize test database: %v", err)
		}
		out.RaceCatIDs[rc.raceCat], err = res.LastInsertId()
		if err != nil {
			t.Fatalf("unable to initialize test database: %v", err)
		}

		for _, gc := range rc.gameCats {
			stmt = fmt.Sprintf("INSERT INTO %s (%s, %s) VALUES (?, ?)", db.TableRaceCatGameCat, db.ColRaceCatGameCatGameCatgeoryID, db.ColRaceCatGameCatRaceCategoryID)

			res, err = database.Exec(stmt, out.GameCatIDs[gc], out.RaceCatIDs[rc.raceCat])
			if err != nil {
				t.Fatalf("unable to initialize test database: %v", err)
			}
		}
	}

	return out
}

//Takes a test and a POST/PATCH handler to call and returns the response decoded as a map
func CallMutationHandler(t *testing.T, test MutationHandlerTest, handlerFunc func(http.ResponseWriter, *http.Request)) map[string]any {
	t.Helper()

	//Encode request body
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(test.Body)
	if err != nil {
		t.Fatalf("%s: failed to encode request: %v", test.TestName, err)
	}

	req := httptest.NewRequest(test.RequestType, test.URL, buf) //Request

	//Call handler and decode response
	res := httptest.NewRecorder() //ResponseRecorder

	//Mux for pattern matching
	mux := http.NewServeMux()
	mux.HandleFunc(test.Pattern, handlerFunc)
	mux.ServeHTTP(res, req)

	var res_map map[string]any
	err = json.NewDecoder(res.Body).Decode(&res_map)
	if err != nil {
		t.Fatalf("%s: failed to decode json response: %v", test.TestName, err)
	}

	if res.Code != test.ExpectedResponseCode {
		t.Errorf("%s: incorrect reponse code. expected %v, got %v", test.TestName, test.ExpectedResponseCode, res.Code)
	}

	success, ok := res_map["success"].(bool)
	if !ok {
		t.Fatalf("%s: unable to parse success field as boolean", test.TestName)
	}

	if success != test.ExpectedSuccess {
		t.Errorf("%s: returned success doesn't match expected value. expected %v, got %v", test.TestName, test.ExpectedSuccess, success)
		
		//Write the error if it can be parsed
		errString, ok := res_map["error"].(string)
		if ok {
			t.Logf("%s", errString)
		}
	}

	//Log response body
	logBody, err := json.MarshalIndent(res_map, "", " ")
	if err == nil {
		t.Logf("%s response: %s", test.TestName, logBody)
	}

	return res_map
}

//Takes a GET test and handler to call and returns the response body
func CallQueryHandler(t *testing.T, test QueryHandlerTest, handlerFunc func(http.ResponseWriter, *http.Request)) map[string]any {
	t.Helper()

	//Build URL and make request
	testURL, err := url.Parse(test.URL)
	if err != nil {
		t.Errorf("%s: error building url: %v", test.TestName, err)
	}

	params := url.Values{}
	for k, v := range test.URLParams {
		for i := range v {
			params.Add(k, v[i])
		}
	}
	testURL.RawQuery = params.Encode()

	req := httptest.NewRequest("GET", testURL.String(), new(bytes.Buffer)) //Request

	//Call handler and decode response
	res := httptest.NewRecorder() //ResponseRecorder

	//Mux for pattern matching
	mux := http.NewServeMux()
	mux.HandleFunc(test.Pattern, handlerFunc)
	mux.ServeHTTP(res, req)

	var res_map map[string]any
	err = json.NewDecoder(res.Body).Decode(&res_map)
	if err != nil {
		t.Fatalf("%s: failed to decode json response: %v", test.TestName, err)
	}

	if res.Code != test.ExpectedResponseCode {
		t.Errorf("%s: incorrect reponse code. expected %v, got %v", test.TestName, test.ExpectedResponseCode, res.Code)
	}

	success, ok := res_map["success"].(bool)
	if !ok {
		t.Fatalf("%s: unable to parse success field as boolean", test.TestName)
	}

	if success != test.ExpectedSuccess {
		t.Errorf("%s: returned success doesn't match expected value. expected %v, got %v", test.TestName, test.ExpectedSuccess, success)
		
		//Write the error if it can be parsed
		errString, ok := res_map["error"].(string)
		if ok {
			t.Logf("%s", errString)
		}
	}

	//Log response body
	logBody, err := json.MarshalIndent(res_map, "", " ")
	if err == nil {
		t.Logf("%s response: %s", test.TestName, logBody)
	}

	return res_map
}

//Helper to confirm that GET dates are within the timeframe
func GetDateBounds(befores []string, afters []string) (string, string) {
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