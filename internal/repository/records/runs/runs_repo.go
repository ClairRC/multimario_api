package runs

import (
	"database/sql"

	"github.com/multimario_api/internal/repository"
	"github.com/multimario_api/internal/repository/gamecategories"
)

type Run struct {
	Category *gamecategories.GameCategory
	Time *repository.NullableStr
	Estimate *repository.NullableStr
}

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