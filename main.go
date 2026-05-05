package main

/*
* This package is to run the server and handle the API calls to the multimario backend database.
 */

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/multimario_api/internal/db"
	"github.com/multimario_api/internal/req_handler"
	"github.com/multimario_api/internal/routes"
	"github.com/multimario_api/internal/twitch"
	_ "github.com/ncruces/go-sqlite3/driver"
)

type Settings struct {
	TwitchClientID string `json:"twitch_client_id"`
	TwitchClientSecret string `json:"twitch_client_secret"`
	DBPath string `json:"database_path"`
}

const port = ":8080" //Port the server listens on

func main() {
	//Load settings
	settings, err := loadSettings("settings.json")
	if err != nil {
		log.Fatal(err)
	}

	//Set Twitch parameters
	err = twitch.SetTwitchParams(settings.TwitchClientID, settings.TwitchClientSecret)
	if err != nil {
		log.Fatal(err)
	}

	//Open database
	database, err := sql.Open("sqlite3", settings.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close() //Defer close database

	//Initialize database if it isn't already
	err = db.DatabaseInit(database)
	if err != nil {
		log.Fatal(err)
	}
	
	mux := http.NewServeMux() //Server Mux for routing
	handler := &req_handler.ReqHandler{DataBase: database} //Make req_handler

	routes.Register(mux, handler) //Register routes

	//Start HTTP server
	log.Fatal(http.ListenAndServe(port, mux))
}

func loadSettings(settingsPath string) (*Settings, error) {
	//Load settings
	settingsFile, err := os.Open(settingsPath)
	if err != nil {
		return nil, err
	}
	defer settingsFile.Close()

	var settings Settings
	err = json.NewDecoder(settingsFile).Decode(&settings)
	if err != nil {
		return nil, err
	}

	return &settings, nil
}