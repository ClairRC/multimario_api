package req_handler

import "net/http"

/*
* Add new Run
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
*	id: int //Run ID
		or
	error: string //Error (only if success is false)
* }
*
*/

func (h *ReqHandler) AddPlayer(w http.ResponseWriter, r *http.Request) {
	//TODO: Implement
}

/*
* Edit existing player
*
* ENDPOINT: PATCH /players/{player_identifier}
*
* EXPECTED:
* {
*	player_name: string //OPTIONAL: Update player's name
*	twitch_id: int //OPTIONAL: Player's twitch ID number
* }
*
* RETURNS:
* {
*	success: boolean //True on successful creation
*	id: int //Run ID
		or
	error: string //Error (only if success is false)
* }
*
*/

func (h *ReqHandler) EditPlayer(w http.ResponseWriter, r *http.Request) {
	//TODO: Implement
}

/*
* Get players
*
* ENDPOINT: GET /players
*
* OPTIONAL PARAMETERS:
*	player_id: int //Returns player with this ID
*	player_name: string //Returns player(s?) with this name
*	twitch_name: string //Returns palyer with this name on twitch
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
* }
*
*/

func (h *ReqHandler) GetPlayers(w http.ResponseWriter, r *http.Request) {
	//TODO: Implement
}