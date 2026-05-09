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
	gameName, err := validateText(w, req, "name", true)
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
	newName, err := validateText(w, req, "name", true)
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
* Get list of games
*
* ENDPOINT: GET /games
*
* OPTIONAL PARAMETERS
* name: string //Name of the game
*
* RETURNS
* {
*	games: //Array of games if success is true
*	[ 
*		{
*			name: string //Game name
*		}
*	]
*	success: boolean
*	error: string //If success is false
* }
*/

func (h *ReqHandler) GetGames(w http.ResponseWriter, r *http.Request) {
	//Get URL names
	gameNames := r.URL.Query()["player_name"]

	//Build query
	q := games.GameQuery{Names: gameNames}

	//Get games
	g, err := games.QueryGames(h.DataBase, q)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unknown error finding games")
	}

	//Set up output
	out := make(map[string]any)
	outGames := make([]map[string]any, 0)
	for _, game := range g {
		newGame := map[string]any {
			"name": game.Name.NullableValue(),
		}
		outGames = append(outGames, newGame)
	}

	out["games"] = outGames
	out["success"] = true

	writeJSON(w, http.StatusOK, out)
}