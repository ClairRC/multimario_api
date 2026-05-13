package req_handler

import (
	"errors"
	"net/http"

	"github.com/multimario_api/internal/repository"
	"github.com/multimario_api/internal/repository/gamecategories"
	"github.com/multimario_api/internal/repository/racecategories"
)

/*
* Add new Race Category to db
*
* ENDPOINT: POST /racecategories
*
* EXPECTED:
* {
*	name: string //REQUIRED -- New category name
*	game_categories: //REQUIRED -- Array of game categories that are apart of this race category
*	[
*		{
*			game_category_name: string //REQUIRED
*		}
*	]
* }
*
* RETURNS:
* {
*	success: boolean //True on successful creation
	error: string //Error (only if success is false)
* }
*
*/

// Add new race category
func (h *ReqHandler) AddRaceCategory(w http.ResponseWriter, r *http.Request) {
	//Get request
	req, err := parseReqJSON(r) //Parse request into map
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error parsing request") //Write error if unable to parse JSON for some reason
		return
	}

	//Verify race category name and make sure it doesn't alraedy exist
	name, err := validateText(w, req, "name", true)
	if err != nil {
		return
	}

	exists, err := racecategories.RaceCategoryExistsByName(h.DataBase, name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unknown error parsing name")
		return
	}
	if exists {
		writeError(w, http.StatusBadRequest, "race category already exists")
		return
	}

	//Validate game categories
	gameCats, err := validateGameCategories(h, w, req, "game_categories", true)
	if err != nil {
		return
	}

	//Create race category from input
	raceCat, err := racecategories.NewRaceCategory(h.DataBase, name, gameCats)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unknown error creating race category")
		return
	}

	err = raceCat.Add(h.DataBase)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unknown error adding race category")
		return
	}

	//All fields are okay
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

/*
* Edit race category
*
* ENDPOINT: PATCH /racecategories/{race_category_name}
* Note: Game categories get replaced by request, not appended, so be sure to include all of the new categories
*
* EXPECTED:
* {
*	name: string //OPTIONAL -- New Category Name
*	game_categories: //OPTIONAL -- Array of game categories that are apart of this race category
*	[
*		game_category_id: int //REQUIRED
*	]
* }
*
* RETURNS:
* {
*	success: boolean //True on successful creation
	error: string //Error (only if success is false)
* }
*
*/

// Edit race category
func (h *ReqHandler) EditRaceCategory(w http.ResponseWriter, r *http.Request) {
	//Get path parameter and make sure it exists
	origCatName := repository.MakeNullableStr(r.PathValue("race_category_name"))
	exists, err := racecategories.RaceCategoryExistsByName(h.DataBase, origCatName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unknown error parsing requested category")
		return
	}
	if !exists {
		writeError(w, http.StatusBadRequest, "race category does not exist")
		return
	}

	//Get request
	req, err := parseReqJSON(r) //Parse request into map
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error parsing request") //Write error if unable to parse JSON for some reason
		return
	}

	//Validate name
	newCatName, err := validateText(w, req, "race_category_name", false)
	if err != nil {
		return
	}

	//Validate game categories
	gameCats, err := validateGameCategories(h, w, req, "game_categories", false)
	if err != nil {
		return
	}

	//Get race category
	raceCat, err := racecategories.GetRaceCategoryByName(h.DataBase, origCatName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unknown request accessing race category resource")
		return
	}

	//Update Race Category
	err = raceCat.Update(h.DataBase, newCatName, gameCats)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unknown error updating game category")
		return
	}

	//We're updated, write success
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

/*
* Get race categories
*
* ENDPOINT: GET /racecategories
*
* OPTIONAL PARAMETERS:
*	race_category: string //Race categories with these names
*	game: string //Race categories including these games
*	game_category: string //Race categories that include this game categories
*
* RETURNS:
* {
*	success: boolean
*	error: string //Only if success is false
*	race_categories: //Array of categories
*	[
*		{
*			name: string //Name of this race category
*			num_collectibles: int //Number of collectibles in this race category
*			game_categories: //List of game categories
*			[
*				{
*					name: string //Name of the game category
*					game: string //Name of the game belonging to this category
*					num_collectibles: int //Number of collectibles in this category
					estimate: string //Default estimate for this category
*				}
*			]
*		}
*	]
* }
*
 */

// Edit race category
func (h *ReqHandler) GetRaceCategories(w http.ResponseWriter, r *http.Request) {
	//Parse URL params
	raceCategoryNames := r.URL.Query()["race_category"]
	gameNames := r.URL.Query()["game"]
	gameCategoryNames := r.URL.Query()["game_category"]

	q := racecategories.RaceCategoryQuery{
		RaceCategories: raceCategoryNames,
		Games: gameNames,
		GameCategories: gameCategoryNames,
	}

	raceCats, err := racecategories.QueryRaceCategories(h.DataBase, q)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unknown error fetching race categories")
		return
	}

	//Write outputs
	out := make(map[string]any)
	outRaceCats := make([]map[string]any, 0)

	for _, c := range raceCats {
		newRaceCat := make(map[string]any)
		raceCatGameCats := make([]map[string]any, 0) //Game categories

		//Loop through game categories for their information
		totalNumCollectibles := 0 //Sum of game category collectibles
		for _, g := range c.GameCategories {
			newGameCat := make(map[string]any)

			newGameCat["name"] = g.Name.Value
			newGameCat["game"] = g.Game.Name.Value
			newGameCat["num_collectibles"] = g.NumCollectibles.Value
			newGameCat["estimate"] = g.Estimate.NullableValue()

			totalNumCollectibles += g.NumCollectibles.Value

			raceCatGameCats = append(raceCatGameCats, newGameCat)
		}

		//Add race category
		newRaceCat["name"] = c.Name.Value
		newRaceCat["num_collectibles"] = totalNumCollectibles
		newRaceCat["game_categories"] = raceCatGameCats

		outRaceCats = append(outRaceCats, newRaceCat)
	}

	//Format output
	out["success"] = true
	out["race_categories"] = outRaceCats

	writeJSON(w, http.StatusOK, out)
}

// Helper function for parsing array of game categories
func validateGameCategories(h *ReqHandler, w http.ResponseWriter, req map[string]any, arrayKey string, required bool) ([]*gamecategories.GameCategory, error) {
	//Validate game category array
	gameCatNames, err := validateArray(w, req, arrayKey, required)
	if err != nil {
		return nil, err
	}

	//Validate each game category inside the array
	gameCats := make([]*gamecategories.GameCategory, 0) //Array for game categories
	for _, v := range gameCatNames {
		//Make sure that array has maps in it
		arrayObj, ok := v.(map[string]any)
		if !ok {
			writeError(w, http.StatusBadRequest, "game category array must hold objects")
			return nil, errors.New("game category is invalid")
		}

		//Get game category name
		gameCatName, err := validateText(w, arrayObj, "game_category_name", true)
		if err != nil {
			return nil, err
		}

		//Get game category from the name
		//This performs a different SELECT query for each name, so it's kinda inefficient, but trivial for these use cases
		gameCat, err := gamecategories.GetGameCategoryByName(h.DataBase, gameCatName)
		if err != nil {
			if err == gamecategories.GameCategoryDoesNotExistErr {
				writeError(w, http.StatusBadRequest, "game category does not exist")
			} else {
				writeError(w, http.StatusInternalServerError, "unknown error parsing game category")
			}
			return nil, err
		}

		//Add to map of game categories
		gameCats = append(gameCats, gameCat)
	}

	return gameCats, nil
}
