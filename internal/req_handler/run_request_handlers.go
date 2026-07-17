package req_handler

import (
	"net/http"
	"strconv"

	"github.com/multimario_api/internal/repository"
	"github.com/multimario_api/internal/repository/gamecategories"
	"github.com/multimario_api/internal/repository/players"
	"github.com/multimario_api/internal/repository/racecategories"
	"github.com/multimario_api/internal/repository/races"
	"github.com/multimario_api/internal/repository/records"
	"github.com/multimario_api/internal/repository/records/runs"
)

/*
* Edit Run
*
* ENDPOINT: PATCH records/{race_id}/{player_name}/runs/{game_category}
*
* EXPECTED:
* {
*	time: string //OPTIONAL hh:mm:ss -- Time this player got in this run
*	estimate: string //OPTIONAL hh:mm:ss -- Estimate this player has for this particular run
*	run_num: int //OPTIONAL -- The number this run was in the race. Races currently require a particular order, but this is here just in case
* }
*
* RETURNS:
* {
*	success: boolean //True on successful creation
		or
	error: string //Error (only if success is false)
* }
*
*/

func (h *ReqHandler) EditRun(w http.ResponseWriter, r *http.Request) {
	//Get path values
	playerName := repository.MakeNullableStr(r.PathValue("player_name"))
	raceIDStr := r.PathValue("race_id")
	catName := repository.MakeNullableStr(r.PathValue("game_category"))

	//Get request
	req, err := parseReqJSON(r) //Parse request into map
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error parsing request") //Write error if unable to parse JSON for some reason
		return
	}

	//Convert race ID to int
	raceID, err := strconv.Atoi(raceIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "race ID can't be parsed as int")
		return
	}

	//Validate request body
	newTime, err := validateDuration(w, req, "time", false)
	if err != nil {
		return
	}
	newEstimate, err := validateDuration(w, req, "estimate", false)
	if err != nil {
		return
	}
	newRunNum, err := validateNumber(w, req, "run_num", false)
	if err != nil {
		return
	}

	//Get record from values
	record, err := records.GetRecord(h.DataBase, repository.MakeNullableInt(raceID), playerName)
	if err != nil {
		switch err{
		case races.RaceDoesNotExistErr:
			writeError(w, http.StatusBadRequest, "race does not exist")
		case players.PlayerDoesNotExistErr:
			writeError(w, http.StatusBadRequest, "player does not exist")
		case records.RecordDoesNotExistErr:
			writeError(w, http.StatusBadRequest, "no race record for this player in this race")
		default:
			writeError(w, http.StatusInternalServerError, "unknown error parsing race record: "+err.Error())
		}
		return
	}

	//Make sure that this race includes the requested category
	//Get Race
	race, err := races.GetRaceByID(h.DataBase, int64(raceID))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to parse race")
		return
	}

	//Check that this race has this game category
	valid, err := racecategories.RaceCatContainsGameCat(h.DataBase, race.RaceCategory.Name.Value, catName.Value)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to determine if race contains this category")
		return
	}
	if !valid {
		writeError(w, http.StatusBadRequest, "race category " + race.RaceCategory.Name.Value + " does not contain this game category: " + catName.Value)
		return
	}

	//Get run from record
	run, err := runs.GetRunFromRecordID(h.DataBase, record.RecordID, catName)
	if err != nil {
		switch err {
		case runs.RunDoesNotExistErr:
			writeError(w, http.StatusBadRequest, "run does not exist")
		case gamecategories.GameCategoryDoesNotExistErr:
			writeError(w, http.StatusBadRequest, "game category does not exist")
		case runs.RunCategoryInvalid:
			writeError(w, http.StatusBadRequest, "run game category is not in this race category")
		default:
			writeError(w, http.StatusInternalServerError, "unknown error fetching run: "+err.Error())
		}
		return
	}

	//Update run
	err = run.Update(h.DataBase, newTime, newEstimate, newRunNum)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unknown error updating run")
		return
	}

	//Run is updated
	writeJSON(w, http.StatusOK, map[string]any{"success": true}, nil)
}

/*
* Add new Run
*
* ENDPOINT: GET /records/runs
*
* OPTIONAL PARAMETERS:
*	page_num: int //Page num. Defaults to 1
*	player_name: string //Runs by this player
*	game_category: string //Runs of this game category
*	race_id: int //Runs for this race
*
* RETURNS:
* {
	success: boolean 
	error: string //only if success is false
*	runs: //Array of runs
*	[
*		{
*			id: int //Run ID
			player_name: string //Name of player
*			race_id: int //Race ID
*			game_category: string //Game category
*			estimate: string //hh:mm:ss Estimate for this run
*			time: string //hh:mm:ss Time gotten in this run (NULL if run not finished)
*		}
*	]
* }
*
*/

func (h *ReqHandler) GetRuns(w http.ResponseWriter, r *http.Request) {
	//Get URL values
	urlPlayerNames := r.URL.Query()["player_name"]
	urlGameCategories := r.URL.Query()["game_category"]
	urlRaceIDs := r.URL.Query()["race_id"]
	urlPageNum := r.URL.Query()["page_num"]

	//Validate race IDs
	raceIDs := make([]int64, 0)
	for _, id := range urlRaceIDs {
		idNum, err := strconv.Atoi(id)
		if err != nil {
			writeError(w, http.StatusBadRequest, "unable to parse id "+id+" as int")
			return
		}
		raceIDs = append(raceIDs, int64(idNum))
	}

	//For names that are twitch names not display names, replace the name with the display name.
	//TODO: This is many extra DB calls and not really super clean. Worth refactoring
	for i, name := range urlPlayerNames{
		player, err := players.GetPlayerByName(h.DataBase, repository.MakeNullableStr(name))
		if err != nil {
			continue 
		} //No player by this name Or twitch name
		urlPlayerNames[i] = player.Name.Value
	}

	q := runs.RunQuery{
		PlayerNames: urlPlayerNames,
		GameCategories: urlGameCategories,
		RaceIDs: raceIDs,
	}

	//Get runs
	//Get response page number
	pageNum, err := getResponsePageNum(urlPageNum)
	if err != nil {
		writeError(w, http.StatusBadRequest, "page number could not be parsed as int")
		return
	}

	//Query stuff
	limit := 50
	offset := limit * (pageNum - 1)

	//Get runs
	runRecords, count, err := runs.QueryRuns(h.DataBase, q, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unknown error fetching runs: "+err.Error())
		return
	}

	//Get output
	out := make(map[string]any)
	outRuns := make([]map[string]any, 0)

	for _, rr := range runRecords {
		newRun := make(map[string]any)
		newRun["id"] = rr.RunID
		newRun["player_name"] = rr.PlayerName
		newRun["race_id"] = rr.RaceID
		newRun["game_category"] = rr.Category.Name.Value
		newRun["estimate"] = rr.Estimate.Value
		newRun["time"] = rr.Time.NullableValue()

		outRuns = append(outRuns, newRun)
	}

	out["success"] = true
	out["runs"] = outRuns
	
	//Get metadata
	meta := getPaginationMetadata(count, r.URL, pageNum, limit)

	writeJSON(w, http.StatusOK, out, meta)
}