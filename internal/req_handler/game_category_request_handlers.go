package req_handler

import (
	"net/http"

	"github.com/multimario_api/internal/repository"
	"github.com/multimario_api/internal/repository/gamecategories"
	"github.com/multimario_api/internal/repository/games"
)

/*
* Add new Game Category to db
*
* ENDPOINT: POST /gamecategories
*
* Note: Despite needing a game name and category name, category names should be prepended with their game name.
* For instance, sms_any% differentiates it from smo_any%. smo_all_moons. etc.
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
	catName, err := validateText(w, req, "category_name", true)
	if err != nil { return }

	//Validate game name
	gameName, err := validateText(w, req, "game_name", true)
	if err != nil { return }

	//Validate estimate
	estimate, err := validateDuration(w, req, "estimate", true)
	if err != nil { return }

	//Validate collectibles
	numCollectibles, err := validateNumber(w, req, "num_collectibles", true)

	//Check that category doesn't already exist
	catExists, err := gamecategories.GameCategoryExistsByName(h.DataBase, catName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unknown error parsing game category name")
		return
	}
	if catExists {
		writeError(w, http.StatusBadRequest, "category already exists")
		return
	}

	//Check to make sure that game does exist
	gameExists, err := games.GameExistsByName(h.DataBase, gameName)
	if err != nil {
		writeError(w, http.StatusBadRequest, "unknown error parsing game name")
		return
	}
	if !gameExists {
		writeError(w, http.StatusBadRequest, "game does not exist")
		return
	}

	//All fields verified
	newCat, err := gamecategories.NewGameCategory(h.DataBase, catName, estimate, numCollectibles, gameName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unknown error creating category")
		return
	}
	
	err = newCat.Add(h.DataBase)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unknown error adding category")
		return
	}

	//Return success
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
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
	//Get path value
	origCatName := repository.MakeNullableStr(r.PathValue("game_category_name"))

	//Get category
	category, err := gamecategories.GetGameCategoryByName(h.DataBase, origCatName)
	if err != nil {
		if err == gamecategories.GameCategoryDoesNotExistErr {
			writeError(w, http.StatusBadRequest, "game category does not exist")
			return
		} else {
			writeError(w, http.StatusInternalServerError, "unknown error finding game category")
		}
	}

	//Get request
	req, err := parseReqJSON(r) //Parse request into map
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error parsing request") //Write error if unable to parse JSON for some reason
		return
	}

	//Check for new category name and whether it already exists
	newCatName, err := validateText(w, req, "category_name", false)
	if err != nil { return }

	exists, err := gamecategories.GameCategoryExistsByName(h.DataBase, newCatName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unknown error parsing new category name")
		return
	}
	if exists {
		writeError(w, http.StatusBadRequest, "new game category already exists")
		return
	}

	//Check for new game name and whether game exists
	newGameName, err := validateText(w, req, "game_name", false)
	if err != nil { return }

	if newGameName.Valid {
		exists, err = games.GameExistsByName(h.DataBase, newGameName)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "unknown error parsing game name")
			return
		}
		if !exists {
			writeError(w, http.StatusBadRequest, "game does not exist")
			return
		}
	}

	//Check for new estimate
	newEstimate, err := validateDuration(w, req, "estimate", false)
	if err != nil { return }

	//Check for num collectibles
	newNumCollectibles, err := validateNumber(w, req, "num_collectibles", false)
	if err != nil { return }

	//All values validated, update category
	err = category.Update(h.DataBase, newCatName, newEstimate, newNumCollectibles, newGameName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unknown error updating game category")
		return
	}

	//Category updated
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
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
*	success: boolean
*	error: string //Only if success is false
*	game_categories: //Array of game categories
*	[
*		{
*			name: string //Category name
*			id: int //Category id
*			game: string //Game that this category is part of
*			estimate: string //hh:mm:ss, default estimate for this category. May be null/
*			num_collectibles: int //Number of collectibles gotten in this category
*		}
*	]
* }
*
*/

//Get game categories
func (h *ReqHandler) GetGameCategories(w http.ResponseWriter, r *http.Request) {
	//Get URL parameters
	catNames := r.URL.Query()["name"]
	gameNames := r.URL.Query()["game"]
	raceCatNames := r.URL.Query()["race_category"]

	//Query for categories
	q := gamecategories.GameCategoryQuery{
		GameNames: gameNames,
		RaceCategories: raceCatNames,
		Names: catNames,
	}
	gameCats, err := gamecategories.QueryGameCategories(h.DataBase, q)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unknown error fetching game categories: " + err.Error())
		return
	}

	//Format output
	out := make(map[string]any)
	outCats := make([]map[string]any, 0) //Array of categories

	for _, c := range gameCats {
		outObj :=  make(map[string]any)
		outObj["name"] = c.Name.Value
		outObj["id"] = c.CategoryID
		outObj["game"] = c.Game.Name.Value
		outObj["estimate"] = c.Estimate.NullableValue() //Can be NULL
		outObj["num_collectibles"] = c.NumCollectibles.Value

		outCats = append(outCats, outObj)
	}

	out["game_categories"] = outCats
	out["success"] = true

	writeJSON(w, http.StatusOK, out)
}