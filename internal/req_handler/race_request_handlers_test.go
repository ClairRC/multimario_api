package req_handler

import (
	"database/sql"
	"log"
	"testing"

	"github.com/multimario_api/internal/db"
	_ "github.com/ncruces/go-sqlite3/driver"
)

/*
* Tests for race request handlers
 */

//Variables for testing environment
var database *sql.DB
var handler *ReqHandler

//Set up testing environment
func TestMain(m *testing.M) {
	err := initTestDB()
	if err != nil {
		log.Fatal(err)
	}

	handler = &ReqHandler{DataBase: database}

	m.Run()
}

/*
* Test CreateRace()
*/

func TestCreateRace(t *testing.T) {
	//TODO: Implement
}

/*
* Helper funtions for test environment
*/

//Seeds database with test values
func initTestDB() (error) {
	var err error
	database, err = sql.Open("sqlite3", ":memory:")
	//DB can't be opened
	if err != nil {
		return err
	}

	//Init database
	db.DatabaseInit(database)

	//Add test values for this package
	raceCatNames := []string {
		"602",
		"246",
		"sandbox_any%",
		"real_category",
	}

	for i := range raceCatNames {
		database.Exec(`INSERT INTO race_categories (race_category_name)
				VALUES (?)`, raceCatNames[i])
	}

	return nil
}

//Resets database to initial state after each test
func resetDB() {
	//Close database and get a new one. Probably better to just delete it but this is fine for testing
	database.Close()
	err := initTestDB()
	if err != nil {
		log.Fatal("unable to reset database. ending testing")
	}
}