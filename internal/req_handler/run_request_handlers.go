package req_handler

import (
	"net/http"
	"strconv"

	"github.com/multimario_api/internal/repository"
	"github.com/multimario_api/internal/repository/players"
	"github.com/multimario_api/internal/repository/races"
	"github.com/multimario_api/internal/repository/records"
	"github.com/multimario_api/internal/repository/records/runs"
)

/*
* Add new Run
*
* ENDPOINT: PATCH records/{race_id}/{player_name}/runs/{category_name}
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
	catName := repository.MakeNullableStr(r.PathValue("category_name"))

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
		default:
			writeError(w, http.StatusInternalServerError, "unknown error parsing race record")
		}
		return
	}

	//Get run from record
	var run *runs.Run = nil
	for _, r := range record.Runs {
		if r.Category.Name.Value == catName.Value {
			run = r
		}
	}
	if run == nil {
		writeError(w, http.StatusBadRequest, "run with this game category not found")
		return
	}

	//Update run
	err = run.Update(h.DataBase, newTime, newEstimate, newRunNum)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unknown error updating run")
	}

	//Run is updated
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

/*
* Add new Run
*
* ENDPOINT: GET /runs
*
* OPTIONAL PARAMETERS:
*	player_id: int //Runs by this player
*	game_category: string //Runs of this game category
*	race_id: int //Runs for this race
*
* RETURNS:
* {
*	runs: //Array of runs
*	[
*		{
*			id: int //Run ID
*			player_id: int //Player ID
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
	//TODO: Implement
}