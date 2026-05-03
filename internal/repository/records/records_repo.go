package records

import (
	"database/sql"
	"errors"

	"github.com/multimario_api/internal/repository"
	"github.com/multimario_api/internal/repository/players"
	"github.com/multimario_api/internal/repository/races"
	"github.com/multimario_api/internal/repository/records/runs"
)

type Record struct {
	Player *players.Player
	Race *races.Race
	FinishTime repository.NullableStr
	NumCollected repository.NullableInt
	Runs []*runs.Run
}

//Default errors
var RecordDoesNotExistErr error = errors.New("race record does not exist")

//Creates new Record and returns a pointer to it
func NewRecord(database *sql.DB, raceID repository.NullableInt, 
	playerName repository.NullableStr, finishTime repository.NullableStr, numCollected repository.NullableInt, 
	runs []*runs.Run) (*Record, error) {
		//TODO: Implement
		return nil, nil
	}

//Gets record from race ID and player name
func GetRecord(database *sql.DB, raceID repository.NullableInt, playerName repository.NullableStr) (*Record, error) {
	//TODO: Implement
	return nil, nil
}

//Adds race to DB
func (r *Record) Add(database *sql.DB) error {
	//TODO: Implement
	return nil
}

//Updates record
func (r *Record) Update(database *sql.DB, newFinishTime repository.NullableStr, newNumCollected repository.NullableInt) error {
	//TODO: Implement
	return nil
}