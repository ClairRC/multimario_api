package req_handler

import "net/http"

/*
* Add counter
*
* ENDPOINT: POST /counters
*
* EXPECTED:
* {
*	twitch_id: int //Twitch ID of this counter
*	adder_name: string //Name of the player who added the counter
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

func (h *ReqHandler) AddCounter(w http.ResponseWriter, r *http.Request) {
	//TODO: Implement
}

/*
* Add counter
*
* ENDPOINT: GET /counters
*
* OPTIONAL PARAMETERS:
*	twitch_name: string //Counter's twitch name
*	added_by: string //Player who added this counter
*
* RETURNS:
* {
*	twitch_name: string //Counter's twitch name. NULL if counter doesn't exist
*	adder: string //Player who added this counter. NULL if counter doesn't exist
* }
*
*/

func (h *ReqHandler) GetCounters(w http.ResponseWriter, r *http.Request) {
	//TODO: Implement
}