package req_handler

import "net/http"

/*
* Adds a Race Record for the given race
* Can be a record for a current race, an upcoming race, or a past race.
*
* ENDPOINT: POST /records
*
* EXPECTS:
* {
* 	race_id: int //REQUIRED -- ID of the race that this record belongs to
*	player_id: int //REQUIRED -- ID of the player who is associated with this record
*	total_time: String //OPTIONAL -- hh:mm:ss format. The time that the player got in this race. NULL if player did not finish/is ongoing
*	num_collectibles: int //OPTIONAL -- The number of collectibles that this player has/got in the race. Defaults to 0.
* }
*
* RETURNS:
* {
*	success: boolean -- True if record is successfully added, False otherwise
*	id: int -- ID of this record if success is true
*		or
*	error: String -- Error message if success is false
* }
 */

func (h *ReqHandler) CreateRecord(w http.ResponseWriter, r *http.Request) {
	//TODO: Implement
}

/*
* Edits a Race Record for the given race
* Can be a record for a current race, an upcoming race, or a past race.
*
* ENDPOINT: PATCH /records/{record_id}
*
* EXPECTS:
* {
* 	total_time: string //OPTIONAL -- hh:mm:ss format. Updates the total time that this player got in this race
*	num_collectibles: int //OPTIONAL -- Updates the number of collectibles for this record
* }
*
*/

func (h *ReqHandler) UpdateRecord(w http.ResponseWriter, r *http.Request) {
	//TODO: Implement
}

/*
* Gets race records
* 
* OPTIONAL PARAMETERS:
*	player_id: int -- Returns records just from specific players. Can be multiple.
*	category: string -- Returns records just of a specific race category
*	before: string -- YYYY-MM-DD Returns records from races before this date
*	after: string -- YYYY-MM-DD Returns records from races after this date
*	on: string -- YYYY-MM-DD Returns records from races on this date
*	time_lowerthan: string -- hh:mm:ss Returns records less than a certain time
*	time_greaterthan: string -- hh:mm:ss Returns record above a certain time
*
* ENDPOINT: GET /records/
*
* RETURNS
* {
*	records: //Array of race records
*	[
*		{
*			player_id: int //ID of the racer this record belongs to
*			race_id: int //ID of the race this record belongs to
*			time: string //hh:mm:ss The time that was gotten by this player in this race. NULL if unfinished
*			num_collectibles: int //Number of collectibles this player got. If the race was finished, this should be the number of collectibles in the category
*		}
*	]
* }
*
*/

func (h *ReqHandler) GetRaceRecords(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
}

/*
* Gets race records of a specific race
* 
* ENDPOINT: GET /records/{race_id}
*
* OPTIONAL PARAMETERS:
*	time_lowerthan: string //hh:mm:ss Returns time from this race lower than this threshold
*	time_greaterthan: string //hh:mm:ss Returns time from this race higher than this threshold
*
* RETURNS
* {
*	records: //Array of race records
*	[
*		{
*			player_id: int //ID of the racer this record belongs to
*			time: string //hh:mm:ss The time that was gotten by this player in this race. NULL if unfinished
*			num_collectibles: int //Number of collectibles this player got. If the race was finished, this should be the number of collectibles in the category
*		}
*	]
* }
*
*/

func (h *ReqHandler) GetRaceRecordsFromRace(w http.ResponseWriter, r *http.Request) {
	//TODO: Implement
}

/*
* Deletes a race record. Really should only be used if a player like cheated or something
* Note this also deletes all the runs from this record.
* 
* ENDPOINT: DELETE /records
*
* EXPECTS
* {
*	player_id: int //REQUIRED -- ID of the player the record belongs to
*	race_id: int //REQUIRED -- ID of the race this record belongs to
* }
*
*/

func (h *ReqHandler) DeleteRaceRecord(w http.ResponseWriter, r *http.Request) {
	//TODO: Implement
}

