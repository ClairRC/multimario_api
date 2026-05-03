package runs

import (
	"database/sql"
	"errors"

	"github.com/multimario_api/internal/repository"
	"github.com/multimario_api/internal/repository/gamecategories"
)

type Run struct {
	Category *gamecategories.GameCategory
	Time repository.NullableStr
	Estimate repository.NullableStr
	RunID int64 //Run ID. Defaults to 0
}

//Default Errors
var RunDoesNotExistErr error = errors.New("run does not exist")

/*
* Run Constructors
*/

//Returns run pointer and error
func NewRun(database *sql.DB, catName repository.NullableStr, 
	time repository.NullableStr, estimate repository.NullableStr) (*Run, error) {
		//TODO: Implement
		return nil, nil
	}

//Creates run with default fields and returns pointer to it
func NewDefaultRun(database *sql.DB, catName repository.NullableStr) (*Run, error) {
	//TODO: Implement
	return nil, nil
}

/*
* Run Methods
*/

//Adds run to DB with specific record ID since runs must have a record ID to exist
func (r *Run) Add(database *sql.DB, recordID int64) error {
	//TODO: Implement
	return nil
}

/*
* Run Helpers
*/
func GetRunFromID(database *sql.DB, runID int64) (*Run, error) {
	//TODO: Implement
	return nil, nil
}