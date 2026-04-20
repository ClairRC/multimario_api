package req_handler

import (
	"net/http"

	"github.com/multimario_api/internal/db"
)

/*
* Add new Game Category to db
*
* ENDPOINT: POST /gamecategories
*
* EXPECTED:
* {
*	category_name: string //REQUIRED -- New category name
*	game_name: string //REQUIRED -- Name of the game this category belongs to
*	estimate: string //REQUIRED hh:mm:ss -- Default estimate for this game
*	num_collectibles: int //REQUIRED Number of collectibles this category gets
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
func (h *ReqHandler) AddGameCategory(w http.ResponseWriter, r *http.Request) {
	//Input validation
	req, err := parseReqJSON(r) //Parse request into map
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error parsing request") //Write error if unable to parse JSON for some reason
		return
	}

	/*
	* Verify category name
	*/
	var catName string
	err = (&TextField{req["category_name"]}).Validate()
	if err != nil {
		writeError(w, http.StatusBadRequest, "category name must be non-empty string")
		return
	} else {catName = req["category_name"].(string)}

	//Check if category exists
	if db.RecordExists(h.DataBase, db.TableGameCategories, db.ColGameCategoryName, catName) {
		writeError(w, http.StatusBadRequest, "category already exists")
		return
	}

	/*
	* Verify game name
	*/
	var gameName string
	err = (&TextField{req["game_name"]}).Validate()
	if err != nil {
		writeError(w, http.StatusBadRequest, "game name must be a non-empty string")
		return
	} else {gameName = req["game_name"].(string)}

	//Check if game already exists
	if !db.RecordExists(h.DataBase, db.TableGames, db.ColGameName, gameName) {
		writeError(w, http.StatusBadRequest, "game does not exist")
		return
	}

	/*
	* Verify Estimate
	*/
	var estimate string
	err = (&TextField{req["estimate"]}).Validate()
	if err != nil {
		switch err {
		case FieldIsWrongFormatErr:
			writeError(w, http.StatusBadRequest, "time is formatted wrong. must be hh:mm:ss")
			return
		case FieldIsEmptyErr:
			writeError(w, http.StatusBadRequest, "category estimate is required")
			return
		}
	} else {estimate = req["estimate"].(string)}

	/*
	* Verify number of collectbiles
	*/
	var numCollectibles int64
	err = (&IntField{req["num_collectibles"]}).Validate()
	if err != nil {
		writeError(w, http.StatusBadRequest, "num_collectibles must be an int")
		return
	} else {numCollectibles =  req["num_collectibles"].(int64)}

	//All fields validated
	id, err := db.AddNewGameCategory(h.DataBase, catName, gameName, estimate, numCollectibles)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unknown error adding game category")
	}

	//Write return values
	res_map := map[string]any {
		"success": true,
		"id": id,
	}
	writeJSON(w, http.StatusCreated, res_map)
}

/*
* Edit game category
*
* ENDPOINT: PATCH /gamecategories/{game_category_id}
*
* EXPECTED:
* {
*	category_name: string //OPTIONAL -- New category name
*	game_name: string //OPTIONAL -- Name of the game this category belongs to
*	num_collectibles: int //OPTIONAL -- number of collectibles this category gets
*	estimate: string //OPTIONAL hh:mm:ss -- Default estimate for this category
* }
*
* RETURNS:
* {
*	success: boolean //True on successful creation
*	id: int //Game category ID
		or
	error: string //Error (only if success is false)
* }
*
*/

//Edit race category
func (h *ReqHandler) EditGameCategory(w http.ResponseWriter, r *http.Request) {
	//TODO: Implement
}

/*
* Get game categories
*
* ENDPOINT: GET /gamecategories
*
* OPTIONAL PARAMETERS:
*	game: string //Categories including this game
*	race_category: string //Categories that are apart of these race categories
*
*
* RETURNS:
* {
*	game_categories: //Array of game categories
*	[
*		{
*			name: string //Category name
*			id: int //Category id
*			game: string //Game that this category is part of
*			num_collectibles: int //Number of collectibles gotten in this category
*		}
*	]
* }
*
*/

//Edit race category
func (h *ReqHandler) GetGameCategories(w http.ResponseWriter, r *http.Request) {
	//TODO: Implement
}

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

//Add new race category
func (h *ReqHandler) AddRaceCategory(w http.ResponseWriter, r *http.Request) {
	//TODO: Implement
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