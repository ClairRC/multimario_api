package req_handler

import "net/http"

/*
* Add new Run
*
* ENDPOINT: POST /runs
*
* EXPECTED:
* {
*	game_category: string //REQUIRED -- Name of the game category this run is for
*	race_id: int //REQUIRED -- ID of the race this run is for
*	player_id: int //REQUIRED -- ID of the player that this run belongs to
*	time: string //OPTIONAL hh:mm:ss -- Time this player got in this run
*	estimate: string //OPTIONAL hh:mm:ss -- Estimate this player had for this particular run
*	run_num: int //OPTIONAL -- The number this run was in the race. Races currently require a particular order, but this is here just in case
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

func (h *ReqHandler) AddRun(w http.ResponseWriter, r *http.Request) {
	//TODO: Implement
}

/*
* Add new Run
*
* ENDPOINT: PATCH /runs/{run_id}
*
* EXPECTED:
* {
*	game_category: string //OPTIONAL -- Name of the game category this run is for
*	race_id: int //OPTIONAL -- ID of the race this run is for
*	player_id: int //OPTIONAL -- ID of the player that this run belongs to
*	time: string //OPTIONAL hh:mm:ss -- Time this player got in this run
*	estimate: string //OPTIONAL hh:mm:ss -- Estimate this player had for this particular run
*	run_num: int //OPTIONAL -- The number this run was in the race. Races currently require a particular order, but this is here just in case
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

func (h *ReqHandler) EditRun(w http.ResponseWriter, r *http.Request) {
	//TODO: Implement
}

/*
* Add new Run
*
* ENDPOINT: GET /runs
*
* OPTIONAL PARAMETERS:
*	player_id: int //Runs by this player
*	game_category: string //Runs of this game category
*	race_id: int //Runs for this race
*
* RETURNS:
* {
*	runs: //Array of runs
*	[
*		{
*			id: int //Run ID
*			player_id: int //Player ID
*			race_id: int //Race ID
*			game_category: string //Game category
*			estimate: string //hh:mm:ss Estimate for this run
*			time: string //hh:mm:ss Time gotten in this run (NULL if run not finished)
*		}
*	]
* }
*
*/

func (h *ReqHandler) GetRuns(w http.ResponseWriter, r *http.Request) {
	//TODO: Implement
}