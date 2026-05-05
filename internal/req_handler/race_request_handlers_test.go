package req_handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"log"
	"net/http/httptest"
	"testing"

	"github.com/multimario_api/internal/db"
	_ "github.com/ncruces/go-sqlite3/driver"
)

/*
* Tests for race request handlers
 */

//Variables for testing environment
var database *sql.DB
var handler *ReqHandler

//Set up testing environment
func TestMain(m *testing.M) {
	var err error
	database, err = sql.Open("sqlite3", ":memory:")

	//If database can't be opened, end the test
	if err != nil {
		log.Fatal(err)
	}

	//Init database
	db.DatabaseInit(database)

	//Add some dummy data to the database
	raceCatNames := []string {
		"602",
		"246",
		"sandbox_any%",
		"real_category",
	}

	for i := range raceCatNames {
		database.Exec(`INSERT INTO race_categories (race_category_name)
				VALUES (?)`, raceCatNames[i])
	}

	handler = &ReqHandler{DataBase: database}

	m.Run()
}

//Test CreateRace()
//Slices of test values
var successTestValues = []map[string]any {
	{"category": "602"}, //Category valid, no date/status
	{"category": "246", "date": "2000-11-16"}, //Category valid, no status
	{"category": "sandbox_any%", "status": "completed"}, //Category valid, no date
	{"category": "real_category", "date": "2000-12-25", "status": ""}, //Category valid, date valid, status invalid (should maybe be an error?)
	{"category": "602", "status": "in_progress"}, //Category valid, status valid, date empty
	{"category": "real_category", "status": "upcoming"}, //Category and status valid, date empty
	{"category": "602", "":true}, //Category valid, extra stuff
}

var failureTestValues = []map[string]any {
	{"":""}, //invalid category
	{"category": 602}, //invalid category
	{"category": "246", "date": "9/11/2001"}, //valid category invalid date
	{"category": "fake_category"}, //invalid category
	{"category": "real_category", "date": "fake_date"}, //valid category, invalid date
	{"category": "real_category", "status": "fake_status"}, //valid category, invalid status
	{"category": "fake_category", "status": "completed"}, //invalid category, valid status
}

func TestCreateRace(t *testing.T) {
	//Loop through Successful Values
	for i := range successTestValues {
		//Encode request body
		var raw_buf []byte
		buf := bytes.NewBuffer(raw_buf)
		err := json.NewEncoder(buf).Encode(successTestValues[i])
		if err != nil {
			log.Fatalf("failed to encode request: %v", err)
		}

		req := httptest.NewRequest("POST", "/racers", buf) //Request
		t.Logf("Test Request Body: %v", req.Body) //Log request

		//Call handler and decode response
		res := httptest.NewRecorder() //ResponseRecorder
		handler.CreateRace(res, req)
		t.Logf("Test Response Body: %v", res.Body) //Log response

		var res_map map[string]any
		err = json.NewDecoder(res.Body).Decode(&res_map)
		if err != nil {
			log.Fatalf("failed to decode response: %v", err)
		}

		//If success is false, that is a Problem
		if res_map["success"] == false {
			t.Error("Expected success, received failure.")
		}
	}

	//Loop through unsuccessful values
	for i := range failureTestValues {
		//Encode request body
		var raw_buf []byte
		buf := bytes.NewBuffer(raw_buf)
		err := json.NewEncoder(buf).Encode(failureTestValues[i])
		if err != nil {
			log.Fatalf("failed to encode request: %v", err)
		}

		req := httptest.NewRequest("POST", "/racers", buf) //Request
		t.Logf("Test Request Body: %v", req.Body) //Log request

		//Call handler and decode response
		res := httptest.NewRecorder() //ResponseRecorder
		handler.CreateRace(res, req)
		t.Logf("Test Response Body: %v", res.Body) //Log response

		var res_map map[string]any
		err = json.NewDecoder(res.Body).Decode(&res_map)
		if err != nil {
			log.Fatalf("failed to decode response: %v", err)
		}

		//If success is false, that is a Problem
		if res_map["success"] == true {
			t.Error("Expected failure, received success.")
		}
	}
}