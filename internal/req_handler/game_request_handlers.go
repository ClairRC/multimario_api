package req_handler

import (
	"net/http"
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
*	id: int //Race ID
		or
	error: string //Error (only if success is false)
* }
*
*/

//Add game to the DB (basically never used but u never know)
func (h *ReqHandler) AddGame(w http.ResponseWriter, r *http.Request) {
	//TODO: Implement
}

/*
* Add new Game to database
*
* ENDPOINT: PATCH /games
*
* EXPECTED:
* {
*	name: string //Game name
* }
*
* RETURNS:
* {
*	success: boolean //True on successful creation
*	id: int //Race ID
		or
	error: string //Error (only if success is false)
* }
*
*/

func (h *ReqHandler) ChangeGameName(w http.ResponseWriter, r *http.Request) {
	//TODO: Implement
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
*	game_id: int //ID number of this game
*	name: string //Game name
* }
*/

func (h *ReqHandler) GetGame(w http.ResponseWriter, r *http.Request) {
	//TODO: Implement
}
