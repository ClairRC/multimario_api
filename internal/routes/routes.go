package routes

/*
* This package will just hold a Register function to register each route to a handler function.
* This is basically where all the routes will go.
 */

import (
	"net/http"

	"github.com/multimario_api/internal/auth"
	"github.com/multimario_api/internal/req_handler"
)

func Register(m *http.ServeMux, h *req_handler.ReqHandler) {
	//Races Handlers
	m.HandleFunc("POST /races", h.Authenticate(h.CreateRace, auth.AuthAdmin))
	m.HandleFunc("PATCH /races/{race_id}", h.Authenticate(h.UpdateRace, auth.AuthAdmin))
	m.HandleFunc("GET /races", h.GetRaces)
	m.HandleFunc("DELETE /races/{id}", h.Authenticate(h.DeleteRace, auth.AuthAdmin))

	//Special handlers for the race that is currently happening
	//Specifically meant to expose useful behavior during races
	m.HandleFunc("PATCH /currentrace/{player_name}/{category_name}", h.Authenticate(h.UpdatePlayerGameTime, auth.AuthVerified))
	m.HandleFunc("PATCH /currentrace/{player_name}", h.Authenticate(h.SetPlayerCollectibleCount, auth.AuthVerified))
	m.HandleFunc("GET /currentrace/runs", h.GetCurrentRaceRuns)
	m.HandleFunc("GET /currentrace", h.GetCurrentRaceStandings)

	//Race Records Handlers
	m.HandleFunc("POST /records", h.Authenticate(h.CreateRecord, auth.AuthAdmin))
	m.HandleFunc("PATCH /records/{race_id}/{player_name}", h.Authenticate(h.UpdateRecord, auth.AuthAdmin))
	m.HandleFunc("GET /records", h.GetRaceRecords)
	m.HandleFunc("DELETE /records/{race_id}/{player_name}", h.Authenticate(h.DeleteRaceRecord, auth.AuthAdmin))

	//Games Handlers
	m.HandleFunc("POST /games", h.Authenticate(h.AddGame, auth.AuthAdmin))
	m.HandleFunc("PATCH /games/{game_name}", h.Authenticate(h.ChangeGameName, auth.AuthAdmin))
	m.HandleFunc("GET /games", h.GetGames)

	//Game Categories Handlers
	m.HandleFunc("POST /gamecategories", h.Authenticate(h.AddGameCategory, auth.AuthAdmin))
	m.HandleFunc("PATCH /gamecategories/{game_category_name}", h.Authenticate(h.EditGameCategory, auth.AuthAdmin))
	m.HandleFunc("GET /gamecategories", h.GetGameCategories)

	//Race Categories Handlers
	m.HandleFunc("POST /racecategories", h.Authenticate(h.AddRaceCategory, auth.AuthAdmin))
	m.HandleFunc("PATCH /racecategories/{race_category_name}", h.Authenticate(h.EditRaceCategory, auth.AuthAdmin))
	m.HandleFunc("GET /racecategories", h.GetRaceCategories)

	//Runs Handlers
	m.HandleFunc("PATCH /records/{race_id}/{player_name}/runs/{game_category}", h.Authenticate(h.EditRun, auth.AuthAdmin))
	m.HandleFunc("GET /records/runs", h.GetRuns)

	//Players Handlers
	m.HandleFunc("POST /players", h.Authenticate(h.AddPlayer, auth.AuthAdmin))
	m.HandleFunc("PATCH /players/{player_name}", h.Authenticate(h.EditPlayer, auth.AuthAdmin))
	m.HandleFunc("GET /players", h.GetPlayers)

	//Handlers for Auth
	m.HandleFunc("GET /auth/api_key", h.CreateAPIKey)
	m.HandleFunc("GET /auth/twitch/callback", h.TwitchCallback)
}
