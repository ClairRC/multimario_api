package req_handler

import (
	"errors"
	"net/http"
	"slices"

	"github.com/multimario_api/internal/repository"
	"github.com/multimario_api/internal/repository/gamecategories"
	"github.com/multimario_api/internal/repository/races"
	"github.com/multimario_api/internal/repository/records/runs"
)

/*
* Adds a Race Record for the given race/player
* Can be a record for a current race, an upcoming race, or a past race.
*
* This is essentially the "signup" for a race. It binds the player to the race
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
	//Input validation
	req, err := parseReqJSON(r) //Parse request into map
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error parsing request") //Write error if unable to parse JSON for some reason
		return
	}

	//Validate body values
	raceID, err := validateNumber(w, req, "race_id", true)
	if err != nil { return }

	playerName, err := validateText(w, req, "player_name", true)
	if err != nil { return }

	finishTime, err := validateDuration(w, req, "finish_time", false)
	if err != nil { return }

	numCollected, err := validateNumber(w, req, "num_collected", false)
	if err != nil { return }

	//Get race and categories
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

	//TODO: finish

	//Thingy is added, return success
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
	//TODO: Implement
}

/*
* Gets race records
* 
* OPTIONAL PARAMETERS:
*	player_id: int -- Returns records just from specific players. Can be multiple.
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
*	records: //Array of race records
*	[
*		{
*			player_id: int //ID of the racer this record belongs to
*			race_id: int //ID of the race this record belongs to
*			time: string //hh:mm:ss The time that was gotten by this player in this race. NULL if unfinished
*			num_collectibles: int //Number of collectibles this player got. If the race was finished, this should be the number of collectibles in the category
*		}
*	]
* }
*
*/

func (h *ReqHandler) GetRaceRecords(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
}

/*
* Gets race records of a specific race
* 
* ENDPOINT: GET /records/{race_id}
*
* OPTIONAL PARAMETERS:
*	time_lowerthan: string //hh:mm:ss Returns time from this race lower than this threshold
*	time_greaterthan: string //hh:mm:ss Returns time from this race higher than this threshold
*
* RETURNS
* {
*	records: //Array of race records
*	[
*		{
*			player_id: int //ID of the racer this record belongs to
*			time: string //hh:mm:ss The time that was gotten by this player in this race. NULL if unfinished
*			num_collectibles: int //Number of collectibles this player got. If the race was finished, this should be the number of collectibles in the category
*		}
*	]
* }
*
*/

func (h *ReqHandler) GetRaceRecordsFromRace(w http.ResponseWriter, r *http.Request) {
	//TODO: Implement
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

/*
* TODO: Refactor these and fix the bugs that i know are there
*/

//Helper function to get runs list
//Takes slice of items from request, returns slice of Runs or error
func getRuns(h *ReqHandler, w http.ResponseWriter, items []map[string]any, categories []*gamecategories.GameCategory, 
	catNameKey string, catTimeKey string, catEstimateKey string) ([]*runs.Run, error) {
		//Output
		out := make([]*runs.Run, 0)

		//If items is empty, create default runs for each category
		if len(items) == 0 {
			for _, cat := range categories {
				newRun, err := runs.NewDefaultRun(h.DataBase, cat.Name)
				if err != nil { return nil, err }
				out = append(out, newRun)
			}
			return out, nil
		}

		//Get runs from categories
		runs, err := validateRuns(h, w, items, categories, catNameKey, catTimeKey, catEstimateKey)
		if err != nil { return nil, err }

		return runs, nil
	}

//Takes a list of items from request, returns the of the categories included in the request or an error
func validateRuns(h *ReqHandler, w http.ResponseWriter, items []map[string]any, categories []*gamecategories.GameCategory,
	catNameKey string, catTimeKey string, catEstimateKey string) ([]*runs.Run, error) {
		//TODO: Fix error handling. Decide where errors get handled and where they don't.
		
		//Gets expected category names for validation
		expectedCatNames := make([]string, 0)
		for _, category := range categories {
			expectedCatNames = append(expectedCatNames, category.Name.Value)
		}

		//Get names of categories from request and validates their values
		catNames := make([]string, 0)//Slice to save included category names
		for _, item := range items {
			itemCatName, err := validateText(w, item, catNameKey, true)
			if err != nil { return nil, err}

			_, err = validateDuration(w, item, catTimeKey, true)
			if err != nil { return nil, err }

			_, err = validateDuration(w, item, catEstimateKey, true)
			if err != nil { return nil, err }

			catNames = append(catNames, itemCatName.Value)
		}

		//Validate that each category is part of the race
		err := validateCategoryNames(catNames, expectedCatNames)
		if err != nil {
			writeError(w, http.StatusBadRequest, "category is not in this race")
			return nil, err
		}

		//For each expected category, if that category wasn't passed in, make a default run
		out := make([]*runs.Run, len(expectedCatNames))
		for _, catName := range expectedCatNames {
			if !slices.Contains(catNames, catName) {
				newRun, err := runs.NewDefaultRun(h.DataBase, repository.MakeNullableStr(catName))
				if err != nil {
					return nil, err
				}
				out = append(out, newRun)
			}
		}

		//For each category that IS inluded in items, make new run
		for _, item := range items {
			catName := repository.MakeNullableStr(item[catNameKey])
			catTime := repository.MakeNullableStr(item[catTimeKey])
			catEstimate := repository.MakeNullableStr(item[catEstimateKey])

			newRun, err := runs.NewRun(h.DataBase, catName, catTime, catEstimate)
			if err != nil {
				return nil, err
			}
			out = append(out, newRun)
		}  

		return out, nil
	}

//Takes slice of category names from request and slice of expected category names and returns an error
//if categories don't match
func validateCategoryNames(catNames []string, expectedNames []string) error {
	for _, catName := range catNames {
		if !slices.Contains(expectedNames, catName) {
			return errors.New("category not in race")
		}
	}

	return nil
}