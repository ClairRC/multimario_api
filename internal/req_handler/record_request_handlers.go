package req_handler

import (
	"errors"
	"net/http"
	"slices"
	"strconv"

	"github.com/multimario_api/internal/repository"
	"github.com/multimario_api/internal/repository/players"
	"github.com/multimario_api/internal/repository/races"
	"github.com/multimario_api/internal/repository/records"
	"github.com/multimario_api/internal/repository/records/runs"
)

/*
* Adds a Race Record for the given race/player
* Can be a record for a current race, an upcoming race, or a past race.
*
* This is essentially the "signup" for a race. It binds the player to the race
*
* Note: For future reference, while we don't currently allow runs to be done in a different order,
*		run numbers will be in the order of the runs passed into the request. Currently, this
*		value only exists for potential future use, but if this changes, that's worth noting.
*
* ENDPOINT: POST /records
*
* EXPECTS:
* {
* 	race_id: int //REQUIRED -- ID of the race that this record belongs to
*	player_name: int //REQUIRED -- Name of the player who is associated with this record
*	finish_time: String //OPTIONAL -- hh:mm:ss format. The time that the player got in this race. Leave this blank if player did not finish
*	num_collected: int //OPTIONAL -- The number of collectibles that this player has/got in the race. Defaults to 0.
*	runs [ //OPTIONAL -- Information about each run that is part of this race. Unfilled runs default to unfinished run with default category estimate
*		{
*			category_name: string //REQUIRED
*			time: string //OPTIONAL -- hh:mm:ss. Empty if player didn't finish
*			estimate: string //OPTIONAL -- hh:mm:ss. Defaults to default category estimate
*		}
*	]
* }
*
* RETURNS:
* {
*	success: boolean -- True if record is successfully added, False otherwise
*	error: String -- Error message if success is false
* }
 */

func (h *ReqHandler) CreateRecord(w http.ResponseWriter, r *http.Request) {
	/*
	* As a note for this handler: Unlike the RaceCategory handler where the GameCategories
	* must already exist, since Runs can only exist if a corresponding Record exists,
	* the Record.Add method is responsible for actually calling Run.Add. It's different from
	* before, but it's more consistent and also guarantees atomicity so that Runs don't get
	* added without a Record.
	*
	* As a separate-ish note, in this case the handler is validating each Run AND generating default Runs
	* if there isn't a corresponding run in the request. This is fine, but it might be better to put in the Record.Add function.
	*/

	//Input validation
	req, err := parseReqJSON(r) //Parse request into map
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error parsing request") //Write error if unable to parse JSON for some reason
		return
	}

	//Validate body values
	raceID, err := validateNumber(w, req, "race_id", true)
	if err != nil {
		return
	}

	playerName, err := validateText(w, req, "player_name", true)
	if err != nil {
		return
	}

	finishTime, err := validateDuration(w, req, "finish_time", false)
	if err != nil {
		return
	}

	numCollected, err := validateNumber(w, req, "num_collected", false)
	if err != nil {
		return
	}
	if !numCollected.Valid {
		numCollected = repository.MakeNullableInt(0) //Set default value to 0
	}

	//Get race from ID
	race, err := races.GetRaceByID(h.DataBase, int64(raceID.Value))
	if err != nil {
		switch err {
		case races.RaceDoesNotExistErr:
			writeError(w, http.StatusBadRequest, "requested race id is invalid")
		default:
			writeError(w, http.StatusInternalServerError, "unknown error getting race information")
		}
		return
	}

	//Get run values
	recordRuns, err := validateRuns(h, w, req, "runs", race)
	if err != nil {
		return
	}

	//Create new record and add it
	record, err := records.NewRecord(h.DataBase, raceID, playerName, finishTime, numCollected)
	if err != nil {
		switch err {
		case players.PlayerDoesNotExistErr:
			writeError(w, http.StatusBadRequest, "player does not exist")
		default:
			writeError(w, http.StatusInternalServerError, "unknown error creating record")
		}
		return
	}

	err = record.Add(h.DataBase, recordRuns)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "uknown error adding record")
		return
	}

	//Record is added, return success
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

/*
* Edits a Race Record for the given race
* Can be a record for a current race, an upcoming race, or a past race.
*
* NOTE: To update specific run, use /records/runs endpoint
*
* ENDPOINT: PATCH /records/{race_id}/{player_name}
*
* EXPECTS:
* {
* 	finish_time: string //OPTIONAL -- hh:mm:ss format. Updates the total time that this player got in this race
*	num_collected: int //OPTIONAL -- Updates the number of collectibles for this record
* }
*
 */

func (h *ReqHandler) UpdateRecord(w http.ResponseWriter, r *http.Request) {
	//Get path values
	playerName := repository.MakeNullableStr(r.PathValue("player_name"))
	raceIDStr := r.PathValue("race_id")

	//Convert race ID to int
	raceID, err := strconv.Atoi(raceIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "race ID can't be parsed as int")
		return
	}

	//Get request
	req, err := parseReqJSON(r) //Parse request into map
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error parsing request") //Write error if unable to parse JSON for some reason
		return
	}

	//Get record for race/player pair
	record, err := records.GetRecord(h.DataBase, repository.MakeNullableInt(raceID), playerName)
	if err != nil {
		switch err {
		case players.PlayerDoesNotExistErr:
			writeError(w, http.StatusBadRequest, "player does not exist")
		case races.RaceDoesNotExistErr:
			writeError(w, http.StatusBadRequest, "race does not exist")
		case records.RecordDoesNotExistErr:
			writeError(w, http.StatusBadRequest, "race record does not exist")
		default:
			writeError(w, http.StatusInternalServerError, "unknown error getting race record")
		}
		return
	}
	
	//Validate new values
	//TODO: Currently this means there's no way to "unfinish" a race because a NULL input
	//will just do nothing. Will have to fix this ideally
	newFinishTime, err := validateDuration(w, req, "finish_time", false)
	if err != nil { return }

	newNumCollected, err := validateNumber(w, req, "num_collected", false)
	if err != nil { return }

	//Update record with new values
	err = record.Update(h.DataBase, newFinishTime, newNumCollected)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unknown error updating record")
		return
	}

	//Updated, return success
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

/*
* Gets race records
*
* OPTIONAL PARAMETERS:
*	player_name: string -- Returns records just from specific players. Can be multiple.
*	race_id: int -- Returns records for this race
*	category: string -- Returns records just of a specific race category
*	before: string -- YYYY-MM-DD Returns records from races before this date
*	after: string -- YYYY-MM-DD Returns records from races after this date
*	on: string -- YYYY-MM-DD Returns records from races on this date
*	time_lowerthan: string -- hh:mm:ss Returns records less than a certain time
*	time_greaterthan: string -- hh:mm:ss Returns record above a certain time
*
* ENDPOINT: GET /records/
*
* RETURNS
* {
	success: boolean 
	error: string //only if success is false
*	records: //Array of race records
*	[
*		{
*			player_id: int //ID of the racer this record belongs to
*			race_id: int //ID of the race this record belongs to
*			time: string //hh:mm:ss The time that was gotten by this player in this race. NULL if unfinished
*			num_collected: int //Number of collectibles this player got. If the race was finished, this should be the number of collectibles in the category
*		}
*	]
* }
*
 */

func (h *ReqHandler) GetRaceRecords(w http.ResponseWriter, r *http.Request) {
	//Get URL parameters
	urlPlayerNames := r.URL.Query()["player_name"]
	urlRaceIDs := r.URL.Query()["race_id"]
	urlCategories := r.URL.Query()["category"]
	urlBeforeDates := r.URL.Query()["before"]
	urlAfterDates := r.URL.Query()["after"]
	urlOnDates := r.URL.Query()["on"]
	urlTimeLower := r.URL.Query()["time_lowerthan"]
	urlTimeGreater := r.URL.Query()["time_greaterthan"]	

	//Validate inputs
	raceIDs := make([]int64, 0)
	for _, id := range urlRaceIDs {
		idNum, err := strconv.Atoi(id)
		if err != nil {
			writeError(w, http.StatusBadRequest, "unable to parse id "+id+" as int")
			return
		}
		raceIDs = append(raceIDs, int64(idNum))
	}

	//Validate dates
	for _, date := range slices.Concat(urlBeforeDates, urlAfterDates, urlOnDates) {
		err := (&DateField{date}).Validate()
		if err == FieldIsWrongFormatErr {
			writeError(w, http.StatusBadRequest, date+" cannot be parsed as date. must be in yyyy-mm-dd format")
			return
		}
	}

	//Validate times
	for _, time := range slices.Concat(urlTimeLower, urlTimeGreater) {
		err := (&DurationField{time}).Validate()
		if err == FieldIsWrongFormatErr {
			writeError(w, http.StatusBadRequest, time+" cannot be parsed as duration. Must be hh:mm:ss format")
			return
		}
	}

	//Build query
	q := records.RecordQuery{
		PlayerNames: urlPlayerNames,
		RaceIDs: raceIDs, 
		Categories: urlCategories,
		BeforeDates: urlBeforeDates,
		AfterDates: urlAfterDates,
		OnDates: urlOnDates,
		LowerThan: urlTimeLower,
		HigherThan: urlTimeGreater,
	}

	//Get records
	records, err := records.QueryRecord(h.DataBase, q)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unknown error fetching race records")
		return
	}

	//Get output
	out := make(map[string]any)
	outRecords := make([]map[string]any, 0)

	for _, r := range records {
		newRecord := make(map[string]any)
		newRecord["player_id"] = r.Player.Name.Value
		newRecord["race_id"] = r.Race.RaceID
		newRecord["time"] = r.FinishTime.NullableValue()
		newRecord["num_collected"] = r.NumCollected.Value

		outRecords = append(outRecords, newRecord)
	}
	out["success"] = true
	out["records"] = outRecords

	writeJSON(w, http.StatusOK, out)
}

/*
* Deletes a race record. Really should only be used if a player like cheated or something
* Note this also deletes all the runs from this record.
*
* ENDPOINT: DELETE /records
*
* EXPECTS
* {
*	player_id: int //REQUIRED -- ID of the player the record belongs to
*	race_id: int //REQUIRED -- ID of the race this record belongs to
* }
*
 */

func (h *ReqHandler) DeleteRaceRecord(w http.ResponseWriter, r *http.Request) {
	//TODO: Implement
}

// Helper function to get runs list
// Returns slice of Runs or error
func validateRuns(h *ReqHandler, w http.ResponseWriter, req map[string]any, arrayKey string, race *races.Race) ([]*runs.Run, error) {
	//Run map to store {cat_name : run}
	runMap := make(map[string]*runs.Run)
	for _, cat := range race.RaceCategory.GameCategories {
		run, err := runs.NewDefaultRun(h.DataBase, cat.Name)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "unknown error getting default runs")
			return nil, err
		}
		runMap[cat.Name.Value] = run
	} //Fill run map with default runs

	//Validate runs from request
	reqRuns, err := validateArray(w, req, arrayKey, false)
	if err != nil {
		return nil, err
	}

	//If runs is empty, return default runs
	if len(reqRuns) == 0 {
		return getRunSliceFromMap(runMap), nil
	}

	//Validate request run fields
	for _, run := range reqRuns {
		//Make sure it is correct format
		_, ok := run.(map[string]any)
		if !ok {
			writeError(w, http.StatusBadRequest, "at least 1 run in array is formatted wrong")
			return nil, errors.New("runs formatted incorrectly")
		}

		//Validate fields
		catName, err := validateText(w, run.(map[string]any), "category_name", true)
		if err != nil {
			return nil, err
		}

		catTime, err := validateDuration(w, run.(map[string]any), "time", false)
		if err != nil {
			return nil, err
		}

		catEstimate, err := validateDuration(w, run.(map[string]any), "estimate", false)
		if err != nil {
			return nil, err
		}

		//Check that run is valid run in map, and if so overwrite the default run with this one
		if _, exists := runMap[catName.Value]; !exists {
			writeError(w, http.StatusBadRequest, catName.Value+" is not apart of race")
			return nil, errors.New("game category not in race")
		}

		newRun, err := runs.NewRun(h.DataBase, catName, catTime, catEstimate)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "unknown error adding "+catName.Value)
			return nil, err
		}
		runMap[catName.Value] = newRun
	}

	//All runs added to map, return slice
	return getRunSliceFromMap(runMap), nil
}

// Helper to get run slice from map
func getRunSliceFromMap(runMap map[string]*runs.Run) []*runs.Run {
	out := make([]*runs.Run, 0)
	for _, v := range runMap {
		out = append(out, v)
	}
	return out
}
