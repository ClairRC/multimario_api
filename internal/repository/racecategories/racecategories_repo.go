package racecategories

import (
	"database/sql"

	"github.com/multimario_api/internal/repository"
	"github.com/multimario_api/internal/repository/gamecategories"
)

//Race category struct
type RaceCategory struct {

}

// Create new race catgegory instance
func NewRaceCategory(database *sql.DB, name repository.NullableStr, gameCats []*gamecategories.GameCategory) (*RaceCategory, error) {
	//TODO: Implement
	return nil, nil
}

//Get race category
func GetRaceCategoryByName() {
	//TODO Implement
}

//Add race category
func (c *RaceCategory) Add(database *sql.DB) error {
	//TODO: Implement
	return  nil
}

//Update race category
func (c *RaceCategory) Update() {
	//TODO: Implement
}

//Checks if race category exists
func RaceCategoryExistsByName(database *sql.DB, name repository.NullableStr) (bool, error) {
	//TODO: Implement
	return false, nil 
}

