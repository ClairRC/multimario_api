package req_handler

import (
	"net/http"

	"github.com/multimario_api/internal/repository"
	"github.com/multimario_api/internal/repository/games"
)

/*
* Add new Game to database
*
* ENDPOINT: POST /games
*
* EXPECTED:
* {
*	name: string //New game name
* }
*
* RETURNS:
* {
*	success: boolean //True on successful creation
	error: string //Error (only if success is false)
* }
*
*/

//Add game to the DB (basically never used but u never know)
func (h *ReqHandler) AddGame(w http.ResponseWriter, r *http.Request) {
	//Parse request
	req, err := parseReqJSON(r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to parse request")
		return
	}

	//Validate game name
	gameName, err := validateGameName(w, req, "name", true)
	if err != nil { return }

	//Get game and add it
	game, _ := games.NewGame(gameName)
	game.Add(h.DataBase)

	//Write output
	writeJSON(w, http.StatusOK, map[string]any{"succes": true})
}

/*
* Add new Game to database
*
* ENDPOINT: PATCH /games/{game_name}
*
* EXPECTED:
* {
*	name: string //Game name
* }
*
* RETURNS:
* {
*	success: boolean //True on successful update
	error: string //Error (only if success is false)
* }
*
*/

func (h *ReqHandler) ChangeGameName(w http.ResponseWriter, r *http.Request) {
	//Get path value
	gameName := repository.MakeNullableStr(r.PathValue("game_name"))

	//Get game
	game, err := games.GetGameByName(h.DataBase, gameName)
	if err != nil {
		switch err {
		case games.GameDoesNotExistErr:
			writeError(w, http.StatusBadRequest, "requested game does not exist")
		case repository.StringIsNullErr:
			writeError(w, http.StatusBadRequest, "game must be non-null")
		default:
			writeError(w, http.StatusInternalServerError, "unknown error finding game")
		return
		}
	}

	//Get request
	req, err := parseReqJSON(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "unknown error parsing request")
		return
	}

	//Validate new game name
	newName, err := validateGameName(w, req, "name", true)
	if err != nil { return }

	//Check whether new name is already being used
	exists, err := games.GameExistsByName(h.DataBase, newName)
	if err != nil { 
		writeError(w, http.StatusInternalServerError, "unknown error parsing new game name")
		return 
	}
	if exists {
		writeError(w, http.StatusBadRequest, "game name already exists")
		return
	}

	//Fields are all verified
	err = game.Update(h.DataBase, newName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "uknown error updating game name")
	}

	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

/*
* Get game
*
* ENDPOINT: GET /games
*
* OPTIONAL PARAMETERS:
*	name: string //The name of the game you want. Only really needed if you need the game id
*	id: int //The id of the game you want
*
* RETURNS
* {
*	games: //Array of games
*	[ 
*		{
*			game_id: int //ID number of this game
*			name: string //Game name
*		}
*	]
* }
*/

func (h *ReqHandler) GetGames(w http.ResponseWriter, r *http.Request) {
	//TODO: Implement
}

/*
* Validation Helper functions
*/
//Validate's name and returns it as nullable string. Returns an error if fatal error occurs
func validateGameName(w http.ResponseWriter, req map[string]any, key string, required bool) (repository.NullableStr, error) {
	err := (&TextField{req[key]}).Validate()

	//Depending on error, do something different
	if err != nil {
		switch err {
		case FieldIsEmptyErr:
			//If this is required, return an error. Otherwise, return NULLstr
			if required {
				writeError(w, http.StatusBadRequest, "game name is required")
				return repository.NULLStr, err
			} else { 
				return repository.NULLStr, nil 
			}
		
		case FieldIsWrongFormatErr:
			writeError(w, http.StatusBadRequest, "game name must be string")
			return repository.NULLStr, err
		}
	}

	return repository.MakeNullableStr(req[key].(string)), nil
}