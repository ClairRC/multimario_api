package req_handler

/*
* Special request handlers specifically for the current ongoing race
* These will be used by the twitch bot and potentially other counting bots for updating during the current race.
 */

import (
	"net/http"
)

/*
* Edit player's progress for current race
*
* ENDPOINT: PATCH /currentrace/{player_id}
*
* EXPECTED:
* {
*	num_collectibles: int //REQUIRED: Number of collectibles this player's count should be updated to
* }
*
 */

func (h *ReqHandler) SetPlayerCollectibleCount(w http.ResponseWriter, r *http.Request) {
	//TODO: IMPLEMENT
}

/*
* Set player's game time for this race
*
* ENDPOINT: PATCH /currentrace/{player_id}/{game_name}
*
* EXPECTED:
* {
*	time: string //REQUIRED -- hh:mm:ss format. Should be the final time that the player got. Don't use this endpoint if the player didnt finish
* }
*
*/

func (h *ReqHandler) UpdatePlayerGameTime(w http.ResponseWriter, r *http.Request) {
	//TODO: IMPLEMENT
}

/*
* Get current race standings
*
* ENDPOINT: GET /currentrace
*
* RETURNS:
* {
*	race_id: int //ID of the current race
* 	standings: -- Array with player id and that player's standing
*	[
*		{
*			player_id: int -- ID of the player
*			num_collectibles: int -- Number of collectibles this player currently has
*		}
*	]
* }
*
*/

func (h *ReqHandler) GetCurrentRaceStandings(w http.ResponseWriter, r *http.Request) {
	//TODO: Implement
}

/*
* Get race data of a specified player
*
* ENDPOINT: GET /currentrace/{player_id}
*
* RETURNS:
* {
* 	num_collectibles: int -- Current total number of collectibles the player has
*	runs: array -- Returns the runs that the player has. If the run is finished, it will have final time, otherwise it'll be NULL
*	[
*		{
*			game_name: string //Game name
*			estimate: string //hh:mm:ss format for the player's estimate for this game. May be NULL
*			final_time: string //hh:mm:ss format for the player's final time for this game. NULL if the run hasn't been finished
*		}
*	]
* }
*
*/

func (h *ReqHandler) GetPlayerProgress(w http.ResponseWriter, r *http.Request) {
	//TODO: Implement
}