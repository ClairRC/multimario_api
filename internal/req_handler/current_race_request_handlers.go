package req_handler

/*
* Special request handlers specifically for the current ongoing race
* These will be used by the twitch bot and potentially other counting bots for updating during the current race.
 */

import (
	"net/http"

	"github.com/multimario_api/internal/repository"
	"github.com/multimario_api/internal/repository/gamecategories"
	"github.com/multimario_api/internal/repository/players"
	"github.com/multimario_api/internal/repository/racecategories"
	"github.com/multimario_api/internal/repository/races"
	"github.com/multimario_api/internal/repository/records"
	"github.com/multimario_api/internal/repository/records/runs"
)

/*
* Edit player's progress for current race
*
* ENDPOINT: PATCH /currentrace/{player_name}
*
* EXPECTED:
* {
*	num_collected: int //OPTIONAL: Number of collectibles this player's count should be updated to
	delta_collected: int //OPTIONAL: Number to increment this player's collectibles by. num_collected takes precedence
* }
*
* RETURNS:
* {
*	success: true //boolean
*	error: string //Only if success is false
*	num_collected: int //Updated num collected. Only if success is true
* }
*
*/

func (h *ReqHandler) SetPlayerCollectibleCount(w http.ResponseWriter, r *http.Request) {
	//Check to make sure there's a race in progress
	currentRaceID := races.GetCurrentRaceID()
	if currentRaceID == 0 {
		writeError(w, http.StatusBadRequest, "there is no race in progress")
		return
	}

	//Get path values
	playerName := repository.MakeNullableStr(r.PathValue("player_name"))

	//Get request
	req, err := parseReqJSON(r) //Parse request into map
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error parsing request") //Write error if unable to parse JSON for some reason
		return
	}

	//Get record for race/player pair
	record, err := records.GetRecord(h.DataBase, repository.MakeNullableInt(currentRaceID), playerName)
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
	newNumCollected, err := validateNumber(w, req, "num_collected", false)
	if err != nil { return }

	deltaNumCollected, err := validateNumber(w, req, "delta_collected", false)
	if err != nil { return }

	//Update the record
	err = record.Update(h.DataBase, repository.NULLStr, newNumCollected, deltaNumCollected)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unknown error updating record")
		return
	}

	//Write success
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "num_collected": record.NumCollected.Value}, nil)
}

/*
* Set player's game time for this race
*
* ENDPOINT: PATCH /currentrace/{player_name}/{category_name}
*
* EXPECTED:
* {
*	time: string //REQUIRED -- hh:mm:ss format. Should be the final time that the player got. Don't use this endpoint if the player didnt finish
* }
*
* RETURNS:
* {
*	success: true //boolean
*	error: string //Only if success is false
* }
*
*/

func (h *ReqHandler) UpdatePlayerGameTime(w http.ResponseWriter, r *http.Request) {
	//Check to make sure there's a race in progress
	currentRaceID := races.GetCurrentRaceID()
	if currentRaceID == 0 {
		writeError(w, http.StatusBadRequest, "there is no race in progress")
		return
	}

	//Get path values
	playerName := repository.MakeNullableStr(r.PathValue("player_name"))
	catName := repository.MakeNullableStr(r.PathValue("category_name"))

	//Get request
	req, err := parseReqJSON(r) //Parse request into map
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error parsing request") //Write error if unable to parse JSON for some reason
		return
	}

	//Validate request body
	newTime, err := validateDuration(w, req, "time", true)
	if err != nil {
		return
	}

	//Get record from values
	record, err := records.GetRecord(h.DataBase, repository.MakeNullableInt(currentRaceID), playerName)
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
	race, err := races.GetRaceByID(h.DataBase, int64(currentRaceID))
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
		writeError(w, http.StatusBadRequest, "race does not contain this game category")
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
	err = run.Update(h.DataBase, newTime, repository.NULLStr, repository.NULLInt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unknown error updating run")
		return
	}

	//Run is updated
	writeJSON(w, http.StatusOK, map[string]any{"success": true}, nil)
}

/*
* Get current race standings
*
* ENDPOINT: GET /currentrace
*
* OPTIONAL PARAMETERS:
*	page_num: int //Page number. Defaults to 1
*	player_name: string -- Returns records just from specific players. Can be multiple.
*
* RETURNS:
* {
*	success: boolean
*	error: string //Only if success is false
*	race_id: int //ID of the current race
* 	standings: -- Array with player name and their race progress
*	[
*		{
*			player_name: string -- Display name of the player
*			num_collected: int -- Number of collectibles this player currently has
*			time: string //hh:mm:ss -- Should be NULL if player isn't finished, though currently this is not guaranteed
*		}
*	]
* }
*
*/

func (h *ReqHandler) GetCurrentRaceStandings(w http.ResponseWriter, r *http.Request) {
	//Check to make sure there's a race in progress
	currentRaceID := races.GetCurrentRaceID()
	if currentRaceID == 0 {
		writeError(w, http.StatusBadRequest, "there is no race in progress")
		return
	}

	//Get URL parameters
	urlPlayerNames := r.URL.Query()["player_name"]
	urlPageNum := r.URL.Query()["page_num"]

	//For names that are twitch names not display names, replace the name with the display name.
	//TODO: This is many extra DB calls and not really super clean. Worth refactoring
	for i, name := range urlPlayerNames{
		player, err := players.GetPlayerByName(h.DataBase, repository.MakeNullableStr(name))
		if err != nil {
			continue 
		} //No player by this name Or twitch name
		urlPlayerNames[i] = player.Name.Value
	}

	//Build query
	q := records.RecordQuery{
		PlayerNames: urlPlayerNames,
		RaceIDs: []int64{currentRaceID}, 
	}

	//Get records
	records, err := records.QueryRecord(h.DataBase, q)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unknown error fetching race records: "+err.Error())
		return
	}

	//Get output
	out := make(map[string]any)
	outRecords := make([]map[string]any, 0)

	for _, r := range records {
		newRecord := make(map[string]any)
		newRecord["player_name"] = r.Player.Name.Value
		newRecord["time"] = r.FinishTime.NullableValue()
		newRecord["num_collected"] = r.NumCollected.Value

		outRecords = append(outRecords, newRecord)
	}

	//Pagination logic
	pageNum, err := getResponsePageNum(urlPageNum)
	limit := 50
	if err != nil {
		writeError(w, http.StatusBadRequest, "unknown error parsing page number")
		return
	}

	outRecords, meta := paginate(outRecords, r.URL, pageNum, limit)

	out["success"] = true
	out["race_id"] = currentRaceID
	out["standings"] = outRecords

	writeJSON(w, http.StatusOK, out, meta)
}

/*
* Get runs data from this race
*
* ENDPOINT: GET /currentrace/runs
*
* OPTIONAL PARAMETERS:
*	page_num: int //Page num. Defaults to 1
*	player_name: string //Runs by this player
*	game_category: string //Runs of this game category
*	
* RETURNS:
* {
*	success: bool
*	error: string //Only if success is false
*	race_id: int //ID of the current race
* 	runs: array -- Returns the runs that the player has. If the run is finished, it will have final time, otherwise it'll be NULL
*	[
*		{
*			player_name: string //Player this run belongs to
*			category_name: string //Category name
*			estimate: string //hh:mm:ss format for the player's estimate for this game. May be NULL
*			time: string //hh:mm:ss format for the player's final time for this game. NULL if the run hasn't been finished
*		}
*	]
* }
*
*/

func (h *ReqHandler) GetCurrentRaceRuns(w http.ResponseWriter, r *http.Request) {
	//Check to make sure there's a race in progress
	currentRaceID := races.GetCurrentRaceID()
	if currentRaceID == 0 {
		writeError(w, http.StatusBadRequest, "there is no race in progress")
		return
	}
	
	//Get URL values
	urlPlayerNames := r.URL.Query()["player_name"]
	urlGameCategories := r.URL.Query()["game_category"]
	urlPageNum := r.URL.Query()["page_num"]

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
		RaceIDs: []int64{currentRaceID},
	}

	//Get runs
	runs, err := runs.QueryRuns(h.DataBase, q)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unknown error fetching runs: "+err.Error())
		return
	}

	//Get output
	out := make(map[string]any)
	outRuns := make([]map[string]any, 0)

	for _, r := range runs {
		newRun := make(map[string]any)
		newRun["player_name"] = r.PlayerName
		newRun["category_name"] = r.Category.Name.Value
		newRun["estimate"] = r.Estimate.Value
		newRun["time"] = r.Time.NullableValue()

		outRuns = append(outRuns, newRun)
	}

	//Add pagination logic
	pageNum, err := getResponsePageNum(urlPageNum)
	if err != nil {
		writeError(w, http.StatusBadRequest, "page number could not be parsed as int")
		return
	}
	
	//Get metadata
	limit := 50
	outRuns, meta := paginate(outRuns, r.URL, pageNum, limit)

	out["success"] = true
	out["race_id"] = currentRaceID
	out["runs"] = outRuns

	writeJSON(w, http.StatusOK, out, meta)
}