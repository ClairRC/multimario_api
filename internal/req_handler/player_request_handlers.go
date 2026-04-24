package req_handler

import (
	"net/http"

	"github.com/multimario_api/internal/repository"
	"github.com/multimario_api/internal/repository/players"
)

/*
* Add new Player
*
* ENDPOINT: POST /players
*
* EXPECTED:
* {
*	player_name: string //REQUIRED -- Player's name
*	twitch_id: int //REQUIRED-- Player's twitch ID. Required for race participation. Just set to 0 if its like a banned player or something
* }
*
* RETURNS:
* {
*	success: boolean //True on successful creation
	error: string //Error (only if success is false)
* }
*
*/

func (h *ReqHandler) AddPlayer(w http.ResponseWriter, r *http.Request) {
	//Parse request
	req, err := parseReqJSON(r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to parse request")
		return
	}

	//Validate name. If there's an error, return
	name, err := validateName(w, req, "player_name", true)
	if err != nil { return }

	//Check that player isn't already in database
	exists, err := players.PlayerExistsByName(h.DataBase, name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error parsing name")
		return
	}

	//If player already exists, write error
	if exists {
		writeError(w, http.StatusBadRequest, "player already exists")
		return
	}

	//TODO: For now I'll just make twitch ID 0
	twitchID := repository.MakeNullableInt(0)

	//Create player
	p, err := players.NewPlayer(name, twitchID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unknown error parsing name")
		return
	}

	//All fields validated, add player
	err = p.Add(h.DataBase)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unknown error adding player")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

/*
* Edit existing player
*
* ENDPOINT: PATCH /players/{player_name}
*
* EXPECTED:
* {
*	player_name: string //OPTIONAL: Update player's name
*	player_twitch_id: int //OPTIONAL: Update player's twitch ID
* }
*
* RETURNS:
* {
*	success: boolean //True on successful creation
	error: string //Error (only if success is false)
* }
*
*/

func (h *ReqHandler) EditPlayer(w http.ResponseWriter, r *http.Request) {
	//Get path value
	player_name := repository.MakeNullableStr(r.PathValue("player_name"))

	//Get player
	player, err := players.GetPlayerByName(h.DataBase, player_name)
	if err != nil {
		switch err {
		case players.PlayerDoesNotExistErr:
			writeError(w, http.StatusBadRequest, "player does not exist")
		case repository.StringIsNullErr:
			writeError(w, http.StatusBadRequest, "player name is null")
		default:
			writeError(w, http.StatusInternalServerError, "unknown error")
		}
		return
	}

	//Parse request
	req, err := parseReqJSON(r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to parse request")
		return
	}

	//Validate player name and ID
	name, err := validateName(w, req, "player_name", false)
	if err != nil { return }

	id, err := validateID(w, req, "player_twitch_id", false)
	if err != nil { return }

	//Make sure player doesn't exist
	if name.Valid {
		exists, err := players.PlayerExistsByName(h.DataBase, name);
		if err != nil {
			writeError(w, http.StatusInternalServerError, "unknown error")
			return
		}

		if exists {
			writeError(w, http.StatusBadRequest, "new player name is already used")
			return
		}
	}

	//Update player
	err = player.Update(h.DataBase, name, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unknown error updating player")
		return
	}

	//Success
	writeJSON(w, http.StatusOK, map[string]any{"success":true})
}

/*
* Get players
*
* ENDPOINT: GET /players
*
* OPTIONAL PARAMETERS:
*	player_id: int //Returns player with this ID
*	player_name: string //Returns player(s?) with this name
*	twitch_name: string //Returns player with this name on twitch
*
* RETURNS:
* {
*	players: //Array of players
*	[
*		{
*			name: string //Player's name
*			id: int //Player's ID
*			twitch_name: string //Twitch name. NULL if player doesn't have associated Twitch
*		}
*	]
*	success: boolean
*	error: string //Only if success is false
* }
*
*/

func (h *ReqHandler) GetPlayers(w http.ResponseWriter, r *http.Request) {
	//TODO: Implement
}

/*
* Helper functions for input validation
*/

//Validate's name and returns it as nullable string. Returns an error if fatal error occurs
func validateName(w http.ResponseWriter, req map[string]any, key string, required bool) (repository.NullableStr, error) {
	err := (&TextField{req[key]}).Validate()

	//Depending on error, do something different
	if err != nil {
		switch err {
		case FieldIsEmptyErr:
			//If this is required, return an error. Otherwise, return NULLstr
			if required {
				writeError(w, http.StatusBadRequest, "name is required")
				return repository.NULLStr, err
			} else { 
				return repository.NULLStr, nil 
			}
		
		case FieldIsWrongFormatErr:
			writeError(w, http.StatusBadRequest, "name must be string")
			return repository.NULLStr, err
		}
	}

	return repository.MakeNullableStr(req[key].(string)), nil
}

//Validate player ID
func validateID(w http.ResponseWriter, req map[string]any, key string, required bool) (repository.NullableInt, error) {
	err := (&IntField{req[key]}).Validate()

	if err != nil {
		switch err {
		//If this value is required, throw an error, otherwise return null string
		case FieldIsEmptyErr:
			if required {
				writeError(w, http.StatusBadRequest, "player id cannot be empty")
				return repository.NULLInt, err
			} else {
				return repository.NULLInt, nil
			}

		case FieldIsWrongFormatErr:
			writeError(w, http.StatusBadRequest, "ID must be an int")
			return repository.NULLInt, nil
		}
	}

	return repository.MakeNullableInt(req[key].(int)), nil
}