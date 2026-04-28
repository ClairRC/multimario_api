package db

/*
* This file will handle general database operations.
 */

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/ncruces/go-sqlite3/driver"
)

//Constants for table and column names
const (
	TableGameCategories = "game_categories"
	ColGameCategoryID = "game_category_id"
	ColGameCategoryGameID = "game_id"
	ColGameCategoryName = "game_category_name"
	ColGameCategoryEstimate = "estimate"
	ColGameCategoryNumCollectibles = "num_collectibles"

	TableGames = "games"
	ColGameID = "game_id"
	ColGameName = "game_name"

	TableRaceCategories = "race_categories"
	ColRaceCategoryID = "race_category_id"
	ColRaceCategoryName = "race_category_name"

	TablePlayers = "players"
	ColPlayerName = "player_name"
	ColPlayerID = "player_id"

	TableSocials = "socials"
	ColPlatformID = "platform_id"

	TableRaceCatGameCat = "race_cat_game_cat"
	ColRaceCatGameCatRaceCategoryID = "race_category_id"
	ColRaceCatGameCatGameCatgeoryID = "game_category_id"

	TableRaces = "races"
	ColRaceRaceCategoryID = "race_category_id"
	ColRaceID = "race_id"
	ColRaceDate = "date"
	ColRaceStartTime = "start_time"
	ColRaceStatus = "status"
)

//Operator type and default operators
type Operator string

const Equals Operator = "="
const NotEquals Operator = "<>"
const LessThan Operator = "<"
const GreaterThan Operator = ">"
const LessThanEqualTo Operator = "<="
const GreaterThanEqualTo Operator = ">="

//Statement types
type StatementType string

const Insert StatementType = "INSERT"
const Update StatementType = "UPDATE"
const Delete StatementType = "DELETE"
const Select StatementType = "SELECT"

//SQL statement
type SQLStatement struct {
	Stmt string
	Args []any
	Type StatementType
}

//Struct for building the where clause
type WhereCondition struct {
	ColName string
	Op Operator
	Value any
}

// Table initializations
var initStatements = []string {
	//Create Players table
	`
	CREATE TABLE IF NOT EXISTS players(
		player_id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		player_name TEXT NOT NULL UNIQUE
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
			ON DELETE CASCADE
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

//Execute SQL statements
//Returns a slice of insert IDs for any insert statements that exist
func ExecuteStatements(db *sql.DB, statements []SQLStatement) ([]int64, error) {
	insertIDs := make([]int64, 0)

	//Start transaction
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	//Execute each statement
	for _, v := range statements {
		res, err := tx.Exec(v.Stmt, v.Args...)
		if err != nil {
			return nil, err
		}

		//Add ID to return slice if statement is an insert
		if v.Type == Insert {
			id, err := res.LastInsertId()
			if err != nil { 
				return nil, err 
			}

			insertIDs = append(insertIDs, id)
		}
	}

	//Commit changes
	err = tx.Commit()
	if err != nil {
		return nil, err
	} else {
		return insertIDs, err
	}
}

//Execute SQL queries
//Returns map of {column, []values}
func ExecuteQueries(db *sql.DB, statements []SQLStatement) (map[string][]any, error) {
	res := make(map[string][]any)

	//Start transaction
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	//Execute each statement
	for _, v := range statements {
		rows, err := tx.Query(v.Stmt, v.Args...)
		if err != nil {
			return nil, err
		}
		
		//Get columns
		cols, err := rows.Columns()
		if err != nil {
			return nil, err
		}

		//Add rows value to results
		for rows.Next() {
			//Make value slice and pointers to pass into scan
			vals := make([]any, len(cols))
			ptrs := make([]any, len(cols))
			for i := range vals {
				ptrs[i] = &vals[i] //Fill pointer slice with val locations
			}

			//Scan row into vals and add to res map
			err = rows.Scan(ptrs...)
			if err != nil {
				return nil, err
			}

			for i, col := range cols {
				res[col] = append(res[col], vals[i])
			}
		}
		rows.Close() //Close rows
	}

	//Check for error in commit
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	//No error, return
	return res, nil
}

//Builds SQL statement from certain parameters
func BuildSelectStatement(columns []string, table string, where []WhereCondition) SQLStatement {
	args := make([]any, 0)
	stmt := "SELECT"
	for i, v := range columns {
		if i > 0 {
			stmt += ","
		}
		stmt += fmt.Sprintf(" %s", v)
	}

	stmt += fmt.Sprintf(" FROM %s", table)

	for i, w := range where {
		if i == 0 {
			stmt += fmt.Sprintf(" WHERE %s %s ?", w.ColName, w.Op)
		} else {
			stmt += fmt.Sprintf(" AND %s %s ?", w.ColName, w.Op)
		}
		args = append(args, w.Value)
	}

	return SQLStatement{stmt, args, Select}
}

func BuildInsertStatement(columns []string, table string, values []any) SQLStatement{
	args := make([]any, 0)
    
    // Column names
    stmt := fmt.Sprintf("INSERT INTO %s (", table)
    for i, col := range columns {
        if i > 0 {
            stmt += ", "
        }
        stmt += col
    }
    stmt += ") VALUES ("
    
    // Placeholders + args
    for i, val := range values {
        if i > 0 {
            stmt += ", "
        }
        stmt += "?"
        args = append(args, val)
    }
    stmt += ")"
    
    return SQLStatement{stmt, args, Insert}
}

func BuildUpdateStatement(columns[]string, newVals []any, table string, where []WhereCondition) (SQLStatement, error) {
	if len(columns) != len(newVals) {
		return SQLStatement{}, errors.New("Unknown error: columns and values length dont match in update statement")
	}
 
	args := make([]any, 0)
	stmt := fmt.Sprintf("UPDATE %s", table)

	stmt += " SET "
	for i, v := range columns {
		if i > 0 {
			stmt += ","
		}
		stmt += fmt.Sprintf(" %s=?", v)
		args = append(args, newVals[i])
	}

	for i, w := range where {
		if i == 0 {
			stmt += fmt.Sprintf(" WHERE %s %s ?", w.ColName, w.Op)
		} else {
			stmt += fmt.Sprintf(" AND %s %s ?", w.ColName, w.Op)
		}
		args = append(args, w.Value)
	}

	return SQLStatement{stmt, args, Update}, nil
}

func BuildDeleteStatement(table string, where []WhereCondition) SQLStatement {
	args := make([]any, 0)
	stmt := fmt.Sprintf("DELETE FROM %s", table)

	for i, w := range where {
		if i == 0 {
			stmt += fmt.Sprintf(" WHERE %s %s ?", w.ColName, w.Op)
		} else {
			stmt += fmt.Sprintf(" AND %s %s ?", w.ColName, w.Op)
		}
		args = append(args, w.Value)
	}

	return SQLStatement{stmt, args, Delete}
}

//Gets the ON clause to prevent very messy string stuff
func getOnClause(table1 string, table2 string, joinCol1 string, joinCol2 string) string {
	return fmt.Sprintf("%s.%s = %s.%s", table1, joinCol1, table2, joinCol2)
}

//Gets linking table
func JoinTables(table1 string, table2 string, joinCol1 string, joinCol2 string) string {
	on := getOnClause(table1, table2, joinCol1, joinCol2)
	return fmt.Sprintf("%s JOIN %s ON %s", table1, table2, on)
}

//Checks if record exists
func RecordExists(db *sql.DB, tableName string, columnName string, value string) (bool, error){
	stmt := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE %s = ?)", tableName, columnName)
	var exists int
	err := db.QueryRow(stmt, value).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists == 1, nil
}