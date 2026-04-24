package req_handler

import (
	"net/http"

	"github.com/multimario_api/internal/repository"
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

	//Validate category name
	catName, err := validateCategoryName(w, req, "game_category", true)
	if err != nil { return }

	
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
	//TODO Implement
}

/*
* Validation Helper Functions
*/
func validateCategoryName(w http.ResponseWriter, req map[string]any, key string, required bool) (repository.NullableStr, error) {
	err := (&TextField{req[key]}).Validate()

	//Check error
	if err != nil {
		switch err{
		case FieldIsEmptyErr:
			if required {
				writeError(w, http.StatusBadRequest, "category game cannot be empty")
				return repository.NULLStr, err
			} else {
				return repository.NULLStr, nil
			}
		case FieldIsWrongFormatErr:
			writeError(w, http.StatusBadRequest, "category name must be string")
			return repository.NULLStr, err
		} 
	}

	return repository.MakeNullableStr(req[key].(string)), nil
}