package req_handler

import (
	"errors"
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
*	name: string //Game category name
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
}

/*
* Edit game category
*
* ENDPOINT: PATCH /gamecategories/{game_category_name}
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
*	error: string //Error (only if success is false)
* }
*
*/

//Edit game category
func (h *ReqHandler) EditGameCategory(w http.ResponseWriter, r *http.Request) {

}

/*
* Get game categories
*
* ENDPOINT: GET /gamecategories
*
* OPTIONAL PARAMETERS:
*	game: string //Categories including this game
*	race_category: string //Categories that are apart of these race categories
*	name: string //To get a specific category
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

//Get game categories
func (h *ReqHandler) GetGameCategories(w http.ResponseWriter, r *http.Request) {
	//Get url parameters
	u := r.URL
	q := u.Query()

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

/*
* Helper functions for validation
*/

//Return {catName, nil} if valid, {"", err} if invalid
func (h *ReqHandler) validateGameCategoryNamePOST(w http.ResponseWriter, req map[string]any, key string) (string, error) {
	err := (&TextField{req[key]}).Validate()
	var catName string
	if err != nil {
		writeError(w, http.StatusBadRequest, "category name must be non-empty string")
		return "", err
	} else {catName = req[key].(string)}

	//Check if category already exists
	if db.GameCategoryExistsFromName(h.DataBase, catName) {
		writeError(w, http.StatusBadRequest, "category already exists")
		return "", errors.New("category already exists")
	}

	return catName, nil
}

//Returns {gameName, nil} if valid, {"", err} if invalid
func (h *ReqHandler) validateGameCategoryGamePOST(w http.ResponseWriter, req map[string]any, key string) (string, error) {
	err := (&TextField{req[key]}).Validate()
	var gameName string
	if err != nil {
		writeError(w, http.StatusBadRequest, "game name must be a non-empty string")
		return "", err
	} else {gameName = req[key].(string)}

	//Check if game already exists
	if !db.GameExistsFromName(h.DataBase, gameName) {
		writeError(w, http.StatusBadRequest, "game does not exist")
		return "", errors.New("game does not exist")
	}

	return gameName, nil
}

//Returns {estimate, nil} if valid, {"", err} if invalid
func (h *ReqHandler) validateGameCategoryEstimatePOST(w http.ResponseWriter, req map[string]any, key string) (string, error) {
	var estimate string
	err := (&TextField{req[key]}).Validate()
	if err != nil {
		switch err {
		case FieldIsWrongFormatErr:
			writeError(w, http.StatusBadRequest, "time is formatted wrong. must be hh:mm:ss")
			return "", err
		case FieldIsEmptyErr:
			writeError(w, http.StatusBadRequest, "category estimate is required")
			return "", err
		}
	} else {estimate = req[key].(string)}

	return estimate, nil
}

//Returns {num_collectibles, nil} if valid, {-1, err} if invalid
func (h *ReqHandler) validateGameCategoryCollectiblesPOST(w http.ResponseWriter, req map[string]any, key string) (int64, error) {
	var numCollectibles int64
	err := (&IntField{req[key]}).Validate()
	if err != nil {
		writeError(w, http.StatusBadRequest, "num_collectibles must be an int")
		return -1, err
	} else {numCollectibles =  req[key].(int64)}

	return numCollectibles, nil
}

//Returns {true, newCatName, nil} if catName should be updated, {false, "", nil} if not, {false, "", err} if theres a fatal error
func (h *ReqHandler) validateGameCategoryNamePATCH(w http.ResponseWriter, req map[string]any, key string) (bool, string, error) {
	err := (&TextField{req[key]}).Validate() //Check validity of string
	newCatName := ""
	updateCatName := true
	if err != nil {
		switch err {
		//Return an error if new category can't be parsed as string
		case FieldIsWrongFormatErr:
			writeError(w, http.StatusBadRequest, "new category must be a string")
			return false, "", err
		case FieldIsEmptyErr:
			updateCatName = false //Field is empty, don't update the category name
		}
	} else {
		newCatName = req[key].(string) //No error means new category name is valid
	}

	return updateCatName, newCatName, nil
}

//Returns {true, newGameName, nil} if game name should be updated, {false, "", nil} if not, {false, "", err} if there's a fatal error
func (h *ReqHandler) validateGameCategoryGamePATCH(w http.ResponseWriter, req map[string]any, key string) (bool, string, error) {
	err := (&TextField{req[key]}).Validate()
	newGameName := ""
	updateGameName := true
	if err != nil {
		switch err {
		//Return an error if new game can't be parsed as a string
		case FieldIsWrongFormatErr:
			writeError(w, http.StatusBadRequest, "new game must be a string")
			return false, "", err
		case FieldIsEmptyErr:
			updateGameName = false //Field is empty, don't update the game name
		}
	} else {
		newGameName = req[key].(string) //No error means new category name is valid
	}

	//Check to make sure new game exists
	if updateGameName {
		if !db.GameExistsFromName(h.DataBase, newGameName) {
			writeError(w, http.StatusBadRequest, "new game does not exist")
			return false, "", errors.New("new game does not exist")
		}
	}

	return updateGameName, newGameName, nil
}

//Returns {true, newEstimate, nil} if estimate should be updated, {false, "", nil} if not, {false, "", err} if error occured
func (h *ReqHandler) validateGameCategoryEstimatePATCH(w http.ResponseWriter, req map[string]any, key string) (bool, string, error) {
	err := (&TimeField{req[key]}).Validate()
	updateEstimate := true
	newEstimate := ""
	if err != nil {
		switch err {
		case FieldIsWrongFormatErr:
			writeError(w, http.StatusBadRequest, "time must be string in form hh:mm:ss")
			return false, "", err
		case FieldIsEmptyErr:
			updateEstimate = false
		}
	} else {
		newEstimate = req[key].(string)
	}

	return updateEstimate, newEstimate, nil
}

//Returns {true, newCollectibles, nil} if collectibles should be updated, {false, -1, nil} if not, {false, -1, err} if error occurs
func (h *ReqHandler) validateGameCategoryCollectiblesPATCH(w http.ResponseWriter, req map[string]any, key string) (bool, int64, error) {
	err := (&IntField{req[key]}).Validate()
	updateCollectibles := true
	newCollectibles := int64(-1)
	if err != nil {
		switch err {
		case FieldIsWrongFormatErr: 
			writeError(w, http.StatusBadRequest, "num_collectibles must be int")
			return false, -1, err
		case FieldIsEmptyErr:
			updateCollectibles = false
		}
	} else {
		newCollectibles = req[key].(int64)
	}

	return updateCollectibles, newCollectibles, nil
}