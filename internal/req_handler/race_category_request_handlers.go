package req_handler

import (
	"errors"
	"net/http"

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
*			game_category_name: int //REQUIRED
*		}
*	]
* }
*
* RETURNS:
* {
*	success: boolean //True on successful creation
*	id: int //Race category ID
		or
	error: string //Error (only if success is false)
* }
*
*/

//Add new race category
func (h *ReqHandler) AddRaceCategory(w http.ResponseWriter, r *http.Request) {
	//Get request
	req, err := parseReqJSON(r) //Parse request into map
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error parsing request") //Write error if unable to parse JSON for some reason
		return
	}

	//Verify race category name and make sure it doesn't alraedy exist
	name, err := validateText(w, req, "name", true)
	if err != nil { return }
	
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
	gameCats, err := validateGameCategories(h, w, req, "game_categories")
	if err != nil { return }

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
* ENDPOINT: PUT /racecategories/{race_category_id}
*
* EXPECTED:
* {
*	name: string //REQUIRED -- New Category Name
*	game_categories: //REQUIRED -- Array of game categories that are apart of this race category
*	[
*		game_category_id: int //REQUIRED
*	]
* }
*
* RETURNS:
* {
*	success: boolean //True on successful creation
*	id: int //Race category ID
		or
	error: string //Error (only if success is false)
* }
*
*/

//Edit race category
func (h *ReqHandler) EditRaceCategory(w http.ResponseWriter, r *http.Request) {
	//TODO: Implement
}

/*
* Get race categories
*
* ENDPOINT: GET /racecategories
*
* OPTIONAL PARAMETERS:
*	game: string //Race categories including this game
*	game_category: string //Race categories that include this game category
*
* RETURNS:
* {
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
*				}
*			]
*		}
*	]
* }
*
*/

//Edit race category
func (h *ReqHandler) GetRaceCategories(w http.ResponseWriter, r *http.Request) {
	//TODO: Implement
}

//Helper function for parsing array of game categories
func validateGameCategories(h *ReqHandler, w http.ResponseWriter, req map[string]any, arrayKey string) ([]*gamecategories.GameCategory, error) {
	//Validate game category array
	gameCatNames, err := validateArray(w, req, arrayKey, true)
	if err != nil { return nil, err }
	
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
		if err != nil { return nil, err }

		//Get game category from the name
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