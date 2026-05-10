package req_handler

import (
	"net/http"
	"slices"
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
*	date: string, //OPTIONAL - YYYY-MM-DD format - Defaults to current date
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

	//Make sure race status is a valid string
	if !raceStatusIsAllowed(raceStatus.Value) {
		writeError(w, http.StatusBadRequest, "race status is invalid value")
		return
	}

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
	err = race.Add(h.DataBase)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error adding new race")
		return
	}

	//All fields validated and race added
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "id": race.RaceID})
}

/*
* Edit race fields of database
*
* ENDPOINT: PATCH /races/{race_id}
*
* EXPECTED:
* {
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
	newDate, err := validateDate(w, req, "date", false)
	if err != nil { return }

	newStatus, err := validateText(w, req, "date", false)
	if err != nil { return }

	newStart, err := validateTime(w, req, "start_time", false)
	if err != nil { return }

	//Update
	err = race.Update(h.DataBase, newDate, newStart, newStatus)
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
* Get races
* ENDPOINT: GET /races
*
* OPTIONAL PARAMETERS:
*	race_id: int -- IDs of the races you want to get
*	before: String -- YYYY-MM-DD Races before this date
*	after: String -- YYYY-MM-DD Races after this date
*	on: String -- YYYY-MM-DD Races on this date
*	category: String -- 602, 246, sandbox_any%, etc
*	status: String -- must be valid status
*
* Example:
* GET /races?before=2020-11-16&after=2018-12-25&category=602&category=246 (url encoded)
* This returns races that happened after 12/25/2018, before 11/16/2020 and includes races that were 602 or 246
*
* RETURNS:
* {
*	success: bool
*	error: string //Only if success is false
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
	//Get URL parameters
	urlIDs := r.URL.Query()["race_id"]
	urlBeforeDates := r.URL.Query()["before"]
	urlAfterDates := r.URL.Query()["after"]
	urlOnDates := r.URL.Query()["on"]
	urlRaceCats := r.URL.Query()["category"]
	urlRaceStatuses := r.URL.Query()["status"]

	//Validate values
	ids := make([]int64, 0)
	for _, id := range urlIDs {
		idNum, err := strconv.Atoi(id)
		if err != nil {
			writeError(w, http.StatusBadRequest, "unable to parse id "+id+" as int")
			return
		}
		ids = append(ids, int64(idNum))
	}

	//Validate dates
	for _, date := range slices.Concat(urlAfterDates, urlBeforeDates, urlOnDates) {
		err := (&DateField{date}).Validate()
		if err != nil {
			switch err {
			case FieldIsWrongFormatErr:
				writeError(w, http.StatusBadRequest, date+" cannot be parsed as date. must be in yyyy-mm-dd format")
			default:
				writeError(w, http.StatusInternalServerError, "unknown error parsing date fields") //Should not ever happen, but safety first.
			}
			return
		}
	}

	//Validate statuses
	for _, status := range urlRaceStatuses {
		if !raceStatusIsAllowed(status) {
			writeError(w, http.StatusBadRequest, status+" is not a valid race status")
			return
		}
	}

	//Query
	q := &races.RaceQuery{
		IDs: ids,
		BeforeDates: urlBeforeDates,
		AfterDates: urlAfterDates,
		OnDates: urlOnDates,
		Categories: urlRaceCats,
		Statuses: urlRaceStatuses,
	}
	
	//TODO: Finish
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

//Helper to make sure passed in race status is valid
func raceStatusIsAllowed(status string) bool {
	//Allowed race statuses
	allowed := []string {
		"upcoming",
		"completed",
		"in_progress",
	}
	return slices.Contains(allowed, status)
}