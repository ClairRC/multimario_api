package db

/*
* This package will handle database operations.
 */

import (
	"database/sql"

	_ "github.com/ncruces/go-sqlite3/driver"
)

//String literals for querying
const (
	TableGameCategories = "game_categories"
	ColGameCategoryName = "game_category_name"
	TableGames = "games"
	ColGameName = "game_name"
)

// Table initializations
var initStatements = []string {
	//Create Players table
	`
	CREATE TABLE IF NOT EXISTS players(
		player_id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		player_name TEXT NOT NULL
	)
	`,
	
	//Create Games table
	`
	CREATE TABLE IF NOT EXISTS games(
		game_id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		game_name TEXT NOT NULL
	)
	`,

	//Create GameCategories table
	`
	CREATE TABLE IF NOT EXISTS game_categories(
		game_category_id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		game_id INTEGER NOT NULL,
		game_category_name TEXT NOT NULL,
		estimate TEXT,
		num_collectibles INTEGER NOT NULL,
		FOREIGN KEY (game_id) REFERENCES games(game_id)
	)
	`,

	//Create RaceCategories table
	`
	CREATE TABLE IF NOT EXISTS race_categories(
		race_category_id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		race_category_name TEXT NOT NULL
	)
	`,

	//Create RaceCat_GameCat linking table
	`
	CREATE TABLE IF NOT EXISTS race_cat_game_cat(
		race_cat_game_cat_id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		race_category_id INTEGER NOT NULL,
		game_category_id INTEGER NOT NULL,
		FOREIGN KEY (race_category_id) REFERENCES race_categories(race_category_id),
		FOREIGN KEY (game_category_id) REFERENCES game_categories(game_category_id)
	)
	`,

	//Create Races table
	`
	CREATE TABLE IF NOT EXISTS races(
		race_id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		race_category_id INTEGER NOT NULL,
		date TEXT,
		start_time TEXT,
		status TEXT NOT NULL,
		FOREIGN KEY (race_category_id) REFERENCES race_categories(race_category_id)
	)
	`,

	//Create Race_Records table
	`
	CREATE TABLE IF NOT EXISTS race_records(
		race_record_id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		race_id INTEGER NOT NULL,
		player_id INTEGER NOT NULL,
		finish_time TEXT,
		num_collectibles INTEGER NOT NULL,
		FOREIGN KEY (race_id) REFERENCES races(race_id)
			ON DELETE CASCADE,
		FOREIGN KEY (player_id) REFERENCES players(player_id)
	)
	`,

	//Create Runs table
	`
	CREATE TABLE IF NOT EXISTS runs(
		run_id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		race_record_id INTEGER NOT NULL,
		game_category_id INTEGER NOT NULL,
		time TEXT,
		estimate TEXT,
		run_num INTEGER NOT NULL,
		FOREIGN KEY (race_record_id) REFERENCES race_records(race_record_id)
			ON DELETE CASCADE,
		FOREIGN KEY (game_category_id) REFERENCES game_categories(game_category_id)
	)
	`,

	//Create Socials table
	`
	CREATE TABLE IF NOT EXISTS socials(
		social_id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		player_id INTEGER NOT NULL,
		platform_name TEXT NOT NULL,
		platform_id INTEGER NOT NULL,
		handle TEXT NOT NULL,
		last_updated TEXT NOT NULL,
		FOREIGN KEY (player_id) REFERENCES players(player_id)
	)
	`,

	//Create Counter table
	`
	CREATE TABLE IF NOT EXISTS counters(
		counter_id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		counter_name TEXT NOT NULL
	)
	`,

	//Create Player_Counter table
	//Currently, any counter can count for any racer, so this 
	//table will just be a "who added this counter" table, but if 
	//counters ever have a list of players to count for, this table is needed for 1NF
	`
	CREATE TABLE IF NOT EXISTS player_counter(
		player_counter_id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		player_id INTEGER NOT NULL,
		counter_id INTEGER NOT NULL,
		FOREIGN KEY (player_id) REFERENCES players(player_id),
		FOREIGN KEY (counter_id) REFERENCES counters(counter_id)
	)
	`,
}


func DatabaseInit(db *sql.DB) error {
	//Enforce foreign keys
	_, err := db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return err
	}

	//Create transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //Rollback transaction if it doesn't finish. #Atomicity

	//Database initialization statements
	for _, stmt := range initStatements {
		_, err = tx.Exec(stmt)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

//Adds race to database
//Returns new race id on success, -1 on failure
func AddNewRace(db *sql.DB, raceCatID int64, date string, status string) (int64, error) {
	//Execute SQLite statement
	result, err := db.Exec(`
		INSERT INTO races(race_category_id, date, status)
        VALUES (?, ?, ?)
	`, raceCatID, date, status)
	if err != nil {
		return -1, err
	}

	resultID, err := result.LastInsertId()
	if err != nil {
		return -1, err
	}

	return resultID, nil
}

//Adds game category to database
func AddNewGameCategory(db *sql.DB, category_name string, game_name string, estimate string, num_collectibles int64) (int64, error) {
	//Get GameID
	stmt := "SELECT game_id FROM games WHERE game_name = ?"
	var game_id int64
	db.QueryRow(stmt, game_name).Scan(&game_id)

	//Add game category
	stmt = `INSERT INTO game_categories(game_id, game_category_name, estimate, num_collectibles)
		VALUES (?, ?, ?, ?)`
	result, err := db.Exec(stmt, game_id, category_name, estimate, num_collectibles)
	if err != nil {
		return -1, err
	}

	resultID, err := result.LastInsertId()
	if err != nil {
		return -1, err
	}

	return resultID, nil
}

//Returns Category ID given category name. Returns an error if category doesn't exist
func GetRaceCategoryIDFromName(db *sql.DB, name string) (int64, error) {
	var id int64

	//Query db for the id
	//This will take the first category that matches the name
	//There shouldn't be duplicates, but even if there is this should be okay 
	//as long as the duplicates are right
	err := db.QueryRow(`
		SELECT race_category_id FROM race_categories
		WHERE race_category_name = ?
	`, name).Scan(&id)

	if err != nil {
		return -1, err
	}

	return id, nil
}

//Function to check if record exists
func RecordExists(db *sql.DB, table_name string, col_name string, value string) bool {
	stmt := "SELECT EXISTS(SELECT 1 FROM ? WHERE ? = ?)"
	var result int
	db.QueryRow(stmt, col_name, table_name, value).Scan(&result)

	if result == 1 {return true} else {return false}
}
