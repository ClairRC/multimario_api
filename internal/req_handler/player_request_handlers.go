package req_handler

import (
	"net/http"

	"github.com/multimario_api/internal/repository"
	"github.com/multimario_api/internal/repository/players"
)

/*
* TODO: Currently these handlers expect a twitch ID, but this should be a name, not an ID
*	    I will change this once I implement twitch integration
 */

/*
* Add new Player
*
* ENDPOINT: POST /players
*
* EXPECTED:
* {
*	display_name: string //OPTIONAL -- Player's name as saved in the DB and displayed on stream
*	twitch_name: string //REQUIRED-- Player's twitch name. Will also be set to the display name if display name is empty
*									 Set to "DNE" if twitch no longer exists or player is banned or something.
*									 Primarily for preserving historical data, as it is preferred to include a twitch even if the player is banned
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
	name, err := validateText(w, req, "player_name", true)
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
*	display_name: string //OPTIONAL: Update player's display name
*	twitch_name: string //OPTIONAL: Update player's twitch ID. This is ONLY necessary if the user is using a new twitch account
*									The DB stores the twitch ID rather than the name, so this isnt necessary for account name changes.
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
	name, err := validateText(w, req, "player_name", false)
	if err != nil { return }

	id, err := validateNumber(w, req, "player_twitch_id", false)
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
