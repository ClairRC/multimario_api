package req_handler

import (
	"net/http"
)

/*
* Add new Race Category to db
*
* ENDPOINT: POST /racecategories
*
* EXPECTED:
* {
*	name: string //REQUIRED -- New category name
*	game_categories: //REQUIRED -- Array of game categories that are apart of this race category
*	[
*		game_category_id: int //REQUIRED
*	]
* }
*
* RETURNS:
* {
*	success: boolean //True on successful creation
*	id: int //Race category ID
		or
	error: string //Error (only if success is false)
* }
*
*/

//Add new race category
func (h *ReqHandler) AddRaceCategory(w http.ResponseWriter, r *http.Request) {
	//TODO: Implement
}

/*
* Edit race category
*
* ENDPOINT: PUT /racecategories/{race_category_id}
*
* EXPECTED:
* {
*	name: string //REQUIRED -- New Category Name
*	game_categories: //REQUIRED -- Array of game categories that are apart of this race category
*	[
*		game_category_id: int //REQUIRED
*	]
* }
*
* RETURNS:
* {
*	success: boolean //True on successful creation
*	id: int //Race category ID
		or
	error: string //Error (only if success is false)
* }
*
*/

//Edit race category
func (h *ReqHandler) EditRaceCategory(w http.ResponseWriter, r *http.Request) {
	//TODO: Implement
}

/*
* Get race categories
*
* ENDPOINT: GET /racecategories
*
* OPTIONAL PARAMETERS:
*	game: string //Race categories including this game
*	game_category: string //Race categories that include this game category
*
* RETURNS:
* {
*	race_categories: //Array of categories
*	[
*		{
*			name: string //Name of this race category
*			num_collectibles: int //Number of collectibles in this race category
*			game_categories: //List of game categories
*			[
*				{
*					name: string //Name of the game category
*					game: string //Name of the game belonging to this category
*					num_collectibles: int //Number of collectibles in this category
*				}
*			]
*		}
*	]
* }
*
*/

//Edit race category
func (h *ReqHandler) GetRaceCategories(w http.ResponseWriter, r *http.Request) {
	//TODO: Implement
}
