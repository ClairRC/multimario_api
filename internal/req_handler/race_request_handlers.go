package req_handler

import (
	"net/http"
	"strconv"

	"github.com/multimario_api/internal/repository"
	"github.com/multimario_api/internal/repository/racecategories"
	"github.com/multimario_api/internal/repository/races"
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
*	start_time: string //OPTIONAL - Must be hh:mm:ss field
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

// Create race
func (h *ReqHandler) CreateRace(w http.ResponseWriter, r *http.Request) {
	//Input validation
	req, err := parseReqJSON(r) //Parse request into map
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error parsing request") //Write error if unable to parse JSON for some reason
		return
	}

	//Validate values
	raceCatName, err := validateText(w, req, "category", true)
	if err != nil {
		return
	}

	raceDate, err := validateDate(w, req, "date", false)
	if err != nil {
		return
	}

	raceStatus, err := validateText(w, req, "status", false)
	if err != nil {
		return
	}

	if !raceStatus.Valid {
		raceStatus = repository.MakeNullableStr("upcoming")
	} //If status is NULL, give it a default value

	raceStartTime, err := validateTime(w, req, "start_time", false)
	if err != nil {
		return
	}

	//Get race and ID
	race, err := races.NewRace(h.DataBase, raceDate, raceStartTime, raceStatus, raceCatName)
	if err != nil {
		switch err {
		case racecategories.RaceCategoryDoesNotExistErr:
			writeError(w, http.StatusBadRequest, "race category does not exist")
		default:
			writeError(w, http.StatusInternalServerError, "unknown error making race")
		}
		return
	}

	//Add the race
	id, err := race.Add(h.DataBase)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error adding new race")
		return
	}

	//All fields validated
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "id": id})
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
* RETURNS:
* {
*	success: boolean //True on successful creation
*	error: string //Error (only if success is false)
* }
 */

func (h *ReqHandler) UpdateRace(w http.ResponseWriter, r *http.Request) {
	//Get path value
	raceIDPathVal := r.PathValue("race_id")

	//Convert ID to int
	raceID, err := strconv.Atoi(raceIDPathVal)
	if err != nil {
		writeError(w, http.StatusBadRequest, "race id cannot be parsed as int")
		return
	}

	//Get race from ID
	race, err := races.GetRaceByID(h.DataBase, int64(raceID))
	if err != nil {
		switch err {
		case races.RaceDoesNotExistErr:
			writeError(w, http.StatusBadRequest, "race does not exist")
		default:
			writeError(w, http.StatusInternalServerError, "unknown error getting race")
		}
		return
	}

	//Get request
	req, err := parseReqJSON(r) //Parse request into map
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error parsing request") //Write error if unable to parse JSON for some reason
		return
	}

	//Validate request parameters
	newCat, err := validateText(w, req, "category", false)
	if err != nil { return }

	newDate, err := validateDate(w, req, "date", false)
	if err != nil { return }

	newStatus, err := validateText(w, req, "date", false)
	if err != nil { return }

	newStart, err := validateTime(w, req, "start_time", false)
	if err != nil { return }

	//Update
	err = race.Update(h.DataBase, int64(raceID), newDate, newStart, newStatus, newCat)
	if err != nil {
		switch err {
		case racecategories.RaceCategoryDoesNotExistErr:
			writeError(w, http.StatusBadRequest, "new category does not exist")
		default:
			writeError(w, http.StatusInternalServerError, "unknown error updating race info")
		}
	}

	//Updated, write success
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
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
