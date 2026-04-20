package main

/*
* This package is to run the server and handle the API calls to the multimario backend database.
 */

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/multimario_api/internal/db"
	"github.com/multimario_api/internal/req_handler"
	"github.com/multimario_api/internal/routes"
	_ "github.com/ncruces/go-sqlite3/driver"
)

const db_path = "./internal/db/mm_db.db" //Path of SQLite database that holds multimario stuff
const port = ":8080" //Port the server listens

func main() {
	//Open database
	database, err := sql.Open("sqlite3", db_path)
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