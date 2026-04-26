package req_handler

import (
	"net/http"
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