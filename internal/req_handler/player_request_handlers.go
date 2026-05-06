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

	//Validate inputs
	name, err := validateText(w, req, "display_name", false)
	if err != nil { return }

	twitchName, err := validateText(w, req, "twitch_name", true)
	if err != nil { return }

	if !name.Valid {
		name = twitchName
	}

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

	//Verify twitch name isn't in use
	exists, err = players.TwitchInUseByName(h.DataBase, twitchName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error parsing twitch name")
		return
	}

	if exists {
		writeError(w, http.StatusBadRequest, "twitch name is already in use")
		return
	}

	//Create player
	p, err := players.NewPlayer(name, twitchName)
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
*	twitch_name: string //OPTIONAL: Update player's twitch name. This is ONLY necessary if the user is using a new twitch account
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
	name, err := validateText(w, req, "display_name", false)
	if err != nil { return }

	twitchName, err := validateText(w, req, "twitch_name", false)
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

	//Make sure twitch name isn't in use
	if twitchName.Valid {
		exists, err := players.TwitchInUseByName(h.DataBase, twitchName)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "unknown error checking if twitch is in use")
			return
		}

		if exists {
			writeError(w, http.StatusBadRequest, "twitch name is already in use")
			return
		}
	}

	//Update player
	err = player.Update(h.DataBase, name, twitchName)
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
*	player_name: string //Returns player with this display name.
*	twitch_name: string //Returns player with this name on twitch
*
* Note: For multiple values, include them separately
* ie. /players?player_name=expreli&player_name=odme_ will return both expreli and odme
* Additionally, if you include both a player name and a twitch name, it will return players that match both
* ie. /players?player_name=expreli&twitch_name=odme_ will return nothing. This behavior is probably not ideal for now.
*
* RETURNS:
* {
*	players: //Array of players
*	[
*		{
*			name: string //Player's name
*			twitch_name: string //Twitch name. NULL if player doesn't have associated Twitch
*		}
*	]
*	success: boolean
*	error: string //Only if success is false
* }
*
*/

func (h *ReqHandler) GetPlayers(w http.ResponseWriter, r *http.Request) {
	//Get URL parameters
	playerNames := r.URL.Query()["player_name"]
	twitchNames := r.URL.Query()["twitch_name"]

	//Get players from query
	query := players.PlayerQuery{
		Names: playerNames,
		TwitchNames: twitchNames,
	}
	players, err := players.QueryPlayers(h.DataBase, query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unknown error fetching players from db")
		return
	}

	out := make(map[string]any)
	playerInfo := make([]map[string]any, 0)

	for _, p := range players {
		info := map[string]any {
			"name": p.Name.Value,
			"twitch_name": p.TwitchName.Value,
		}
		playerInfo = append(playerInfo, info)
	}

	//Write outputs
	out["players"] = playerInfo
	out["success"] = true

	writeJSON(w, http.StatusOK, out)
}
