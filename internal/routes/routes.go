package routes

/*
* This package will just hold a Register function to register each route to a handler function.
* This is basically where all the routes will go.
 */

import (
	"net/http"

	"github.com/multimario_api/internal/req_handler"
)

func Register(m *http.ServeMux, h *req_handler.ReqHandler) {
	//Races Handlers
	m.HandleFunc("POST /races", h.CreateRace)
	m.HandleFunc("PATCH /races/{race_id}", h.UpdateRace)
	m.HandleFunc("GET /races/{id}", h.GetRaceFromID)
	m.HandleFunc("GET /races", h.GetRaces)
	m.HandleFunc("DELETE /races/{id}", h.DeleteRace)

	//Special handlers for the race that is currently happening
	//Specifically meant to expose useful behavior during races
	m.HandleFunc("PATCH /currentrace/{player_id}/{game_id}", h.UpdatePlayerGameTime)
	m.HandleFunc("PATCH /currentrace/{player_id}", h.SetPlayerCollectibleCount)
	m.HandleFunc("GET /currentrace/{player_id}", h.GetPlayerProgress)
	m.HandleFunc("GET /currentrace", h.GetCurrentRaceStandings)

	//Race Records Handlers
	m.HandleFunc("POST /records", h.CreateRecord)
	m.HandleFunc("PATCH /records", h.UpdateRecord)
	m.HandleFunc("GET /records/{race_id}", h.GetRaceRecordsFromRace)
	m.HandleFunc("GET /records", h.GetRaceRecords)
	m.HandleFunc("DELETE /records", h.DeleteRaceRecord)

	//Games Handlers
	m.HandleFunc("POST /games", h.AddGame)
	m.HandleFunc("PATCH /games/{game_name}", h.ChangeGameName)
	m.HandleFunc("GET /games", h.GetGames)

	//Game Categories Handlers
	m.HandleFunc("POST /gamecategories", h.AddGameCategory)
	m.HandleFunc("PATCH /gamecategories/{game_category_name}", h.EditGameCategory)
	m.HandleFunc("GET /gamecategories", h.GetGameCategories)

	//Race Categories Handlers
	m.HandleFunc("POST /racecategories", h.AddRaceCategory)
	m.HandleFunc("PATCH /racecategories/{race_category_name}", h.EditRaceCategory)
	m.HandleFunc("GET /racecategories", h.GetRaceCategories)

	//Runs Handlers
	m.HandleFunc("POST /runs", h.AddRun)
	m.HandleFunc("PATCH /runs/{run_id}", h.EditRun)
	m.HandleFunc("GET /runs", h.GetRuns)

	//Players Handlers
	m.HandleFunc("POST /players", h.AddPlayer)
	m.HandleFunc("PATCH /players/{player_name}", h.EditPlayer)
	m.HandleFunc("GET /players", h.GetPlayers)

	//Counters Handlers
	m.HandleFunc("POST /counters", h.AddCounter)
	m.HandleFunc("GET /counters", h.GetCounters)
}
