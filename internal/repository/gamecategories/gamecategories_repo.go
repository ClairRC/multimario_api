package gamecategories

import (
	"database/sql"

	"github.com/multimario_api/internal/db"
	"github.com/multimario_api/internal/repository"
	"github.com/multimario_api/internal/repository/games"
)

//Game category struct
type GameCategory struct {
	Name repository.NullableStr
	Estimate repository.NullableStr
	NumCollectibles repository.NullableInt
	GameName repository.NullableStr
}

// Create new game catgegory instance
func NewGameCategory(name repository.NullableStr, 
	estimate repository.NullableStr, 
	numCollectibles repository.NullableInt, 
	gameName repository.NullableStr) (*GameCategory, error) {

	if !name.Valid {
		return nil, repository.StringIsNullErr
	}

	if !estimate.Valid {
		return nil, repository.StringIsNullErr
	}

	if !numCollectibles.Valid {
		return nil, repository.IntIsNullErr
	}

	if !gameName.Valid {
		return nil, repository.StringIsNullErr
	}

	return &GameCategory{
		Name: name,
		Estimate: estimate,
		NumCollectibles: numCollectibles,
		GameName: gameName,
		}, nil
}

//Add game category
func (c *GameCategory) Add(database *sql.DB) error {
	//Get game FK
	gameID, err := games.GetGameIDFromName(database, c.GameName.Value)
	if err != nil {
		return err
	}

	//Build SQL statements
	stmt := db.BuildInsertStatement(
		[]string{
			db.ColGameCategoryName, db.ColGameCategoryEstimate, 
			db.ColGameCategoryNumCollectibles, db.ColGameCategoryGameID}, 
		
		db.TableGameCategories, 
		
		[]any{c.Name.Value, c.Estimate.Value, c.NumCollectibles.Value, gameID},
	)
	
	return db.ExecuteStatements(database, []db.SQLStatement{stmt})
}

//Checks if gamecategory exists
func GameCategoryExistsByName(database *sql.DB, name repository.NullableStr) (bool, error) {
	if !name.Valid {
		return false, repository.StringIsNullErr
	}

	exists, err := db.RecordExists(database, db.TableGameCategories, db.ColGameCategoryName, name.Value)
	if err != nil {
		return false, err
	}

	return exists, nil
}