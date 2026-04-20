package req_handler

import (
	"net/http"
	"time"

	"github.com/multimario_api/internal/db"
)

/*
* Add new Race to database
*
* ENDPOINT: POST /races
*
* EXPECTED:
* {
*	category: string, //REQUIRED - The category of the race
*	date: string, //OPTIONAL - YYYY-MM-DD format
*	status: string //OPTIONAL - Defaults to "upcoming"
* }
*
* RETURNS:
* {
*	success: boolean //True on successful creation
*	id: int //Race ID
		or
	error: string //Error (only if success is false)
* }
*
*/

//Create race
func (h *ReqHandler) CreateRace(w http.ResponseWriter, r *http.Request) {
	//Input validation
	req, err := parseReqJSON(r) //Parse request into map
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error parsing request") //Write error if unable to parse JSON for some reason
		return
	}

	/*
	* Check if category is valid
	*/
	var catName string
	err = (&RaceCategoryField{req["category"]}).Validate()
	if err != nil {
		writeError(w, http.StatusBadRequest, "category must be a non-empty string")
		return
	} else {
		catName = req["category"].(string)
	}

	//Get category ID
	catID, err := db.GetRaceCategoryIDFromName(h.DataBase, catName)
	if err != nil{
		writeError(w, http.StatusBadRequest, "race category does not exist")
		return
	}

	/*
	* Validate Date field
	*/
	var date string
	err = (&DateField{req["date"]}).Validate()
	if err != nil {
		//If date is in wrong format, send an error, if it's just empty, set it to default valud
		switch err { 
		case FieldIsWrongFormatErr: //Return and write error if field is in wrong formate
			writeError(w, http.StatusBadRequest, "error parsing date. make sure date is in YYYY-MM-DD format")
			return
		case FieldIsEmptyErr:
			date = time.Now().Local().Format(time.DateOnly) //Set date to current date
		default:
			writeError(w, http.StatusInternalServerError, "unknown error parsing date field")
			return
		}
	} else {
		date = req["date"].(string)
	}

	/*
	* Check if status is valid
	*/
	var status string
	err = (&RaceStatusField{req["status"]}).Validate()
	if err != nil {
		switch err {
		case FieldIsInvalidErr:
			writeError(w, http.StatusBadRequest, "invalid status")
			return
		case FieldIsWrongFormatErr:
			writeError(w, http.StatusBadRequest, "status must be a string")
			return
		case FieldIsEmptyErr:
			status = "upcoming"
		default:
			writeError(w, http.StatusInternalServerError, "unknown error parsing race status")
			return
		}
	} else {
		status = req["status"].(string)
	}

	//Add to database
	raceID, err := db.AddNewRace(h.DataBase, catID, date, status)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error adding race to database")
		return
	}

	//Write return value
	res := map[string]any {
		"success": true,
		"id": raceID,
	}

	writeJSON(w, http.StatusCreated, res)
}

/*
* Edit race fields of database
*
* ENDPOINT: PATCH /races/{race_id}
*
* EXPECTED:
* {
*	category: string -- OPTIONAL //Update the category being run
*	date: string -- OPTIONAL //Update the date of the race
*	status: string -- OPTIONAL //Update the status of the race
*	start_time: string -- OPTIONAL //Update the start time
* }
*
*/

func (h *ReqHandler) UpdateRace(w http.ResponseWriter, r *http.Request) {
	//TODO: Implement
}

/*
* Get specific race from database
*
* ENDPOINT: GET /races/{race_id}
*
* RETURNS:
* {
*	id: int -- Race ID
*	category: String
*	date: String -- YYYY-MM-DD
*	status: String
*	start_time: string -- UTC
* }
*/

func (h *ReqHandler) GetRaceFromID(w http.ResponseWriter, r *http.Request) {
	//TODO: Implement
}

/*
* Get races
* ENDPOINT: GET /races
*
* OPTIONAL PARAMETERS:
*	before: String -- YYYY-MM-DD Races before this date
*	after: String -- YYYY-MM-DD Races after this date
*	on: String -- YYYY-MM-DD Races on this date
*	category: String -- 602, 246, sandbox_any%, etc
*
* RETURNS:
* {
* 	races: Array of races
*   [
*		{
*			id: int -- Race ID
*			category: String
*			date: String -- YYYY-MM-DD
*			status: String
*			start_time: string -- UTC
*		}
*	]
* }
*/

func (h *ReqHandler) GetRaces(w http.ResponseWriter, r *http.Request) {
	//TODO: Implement
}

/*
* Delete a race
* Note: This also deletes all records associated with this race.
*
* ENDPOINT: DELETE /races/{id}
*
* EXPECTED: 
* {
* 	id: int //REQUIRED -- Race ID to delete
* }
*/

func (h *ReqHandler) DeleteRace(w http.ResponseWriter, r *http.Request) {
	//TODO: Implement
}