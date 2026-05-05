package gamecategories

import (
	"database/sql"
	"errors"

	"github.com/multimario_api/internal/db"
	"github.com/multimario_api/internal/repository"
	"github.com/multimario_api/internal/repository/games"
)

// Game category struct
type GameCategory struct {
	Name repository.NullableStr
	Estimate repository.NullableStr
	NumCollectibles repository.NullableInt
	Game *games.Game
	CategoryID int64 //0 if category hasn't been added to database yet
}

// Errors
var GameCategoryDoesNotExistErr error = errors.New("game category does not exist")

/*
* GameCategory Constructor
*/

// Create new game category instance
func NewGameCategory(database *sql.DB, name repository.NullableStr,
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

	//Get Game from the name
	game, err := games.GetGameByName(database, gameName)
	if err != nil {
		return nil, err
	}

	return &GameCategory{
		Name:            name,
		Estimate:        estimate,
		NumCollectibles: numCollectibles,
		Game:        game,
	}, nil
}

/*
* GameCategory Methods
*/

// Add game category
func (c *GameCategory) Add(database *sql.DB) error {
	if c.CategoryID != 0 {
		return errors.New("game category already exists")
	}

	//Get game FK
	gameID := c.Game.GameID
	if gameID == 0 {
		return games.GameDoesNotExistErr
	}

	//Build SQL statements
	stmt := db.BuildInsertStatement(
		[]string{
			db.ColGameCategoryName, db.ColGameCategoryEstimate,
			db.ColGameCategoryNumCollectibles, db.ColGameCategoryGameID},

		db.TableGameCategories,

		[]any{c.Name.Value, c.Estimate.Value, c.NumCollectibles.Value, gameID},
	)

	ids, err := db.ExecuteStatements(database, []db.SQLStatement{stmt})
	if err != nil {
		return err
	}

	c.CategoryID = ids[0]
	return nil
}

// Update game category
func (c *GameCategory) Update(
	database *sql.DB, newName repository.NullableStr,
	newEstimate repository.NullableStr, newNumCollectibles repository.NullableInt,
	newGameName repository.NullableStr,
) error {
	//Check that game category is in DB
	if c.CategoryID == 0 {
		return GameCategoryDoesNotExistErr
	}

	//Cols to update
	cols := make([]string, 0)
	newVals := make([]any, 0)

	//Check each field
	if newName.Valid {
		cols = append(cols, db.ColGameCategoryName)
		newVals = append(newVals, newName.Value)
	}

	if newEstimate.Valid {
		cols = append(cols, db.ColGameCategoryEstimate)
		newVals = append(newVals, newEstimate.Value)
	}

	if newNumCollectibles.Valid {
		cols = append(cols, db.ColGameCategoryNumCollectibles)
		newVals = append(newVals, newNumCollectibles.Value)
	}

	var newGame *games.Game = nil //New game to update passed in struct
	if newGameName.Valid {
		//Get game ID from name
		game, err := games.GetGameByName(database, newGameName)
		if err != nil {
			return err
		} //Return if there's an error getting game

		newGame = game
		cols = append(cols, db.ColGameCategoryGameID)
		newVals = append(newVals, game.GameID)
	}

	//If there's nothing new to update, just return
	if len(cols) == 0 {
		return nil
	}

	//Otherwise, build and execute statement
	whereCon := db.WhereCondition{
		ColName: db.ColGameCategoryID,
		Op:      db.Equals,
		Value:   c.CategoryID,
	}

	stmt, err := db.BuildUpdateStatement(cols, newVals, db.TableGameCategories, []db.WhereCondition{whereCon})
	if err != nil {
		return err
	} //Return error if can't build statement

	_, err = db.ExecuteStatements(database, []db.SQLStatement{stmt})
	if err != nil {
		return err
	}

	//No errors, make sure c.Game is updated
	if newGame != nil {
		c.Game = newGame
	}
	
	return nil
}

/*
* Helpers
*/

// Get game category
func GetGameCategoryByName(database *sql.DB, name repository.NullableStr) (*GameCategory, error) {
	if !name.Valid {
		return nil, repository.StringIsNullErr
	}

	//Query database for this game category
	cols := []string{
		db.ColGameCategoryName,
		db.ColGameCategoryEstimate,
		db.ColGameCategoryGameID,
		db.ColGameCategoryNumCollectibles,
		db.ColGameCategoryID,
	}
	table := db.TableGameCategories
	where := []db.WhereCondition{
		{ColName: db.ColGameCategoryName, Op: db.Equals, Value: name.Value},
	}

	stmt := db.BuildSelectStatement(cols, table, where)
	res, err := db.ExecuteQueries(database, []db.SQLStatement{stmt})
	if err != nil {
		return nil, err
	}

	//Get game category from the query. If there's none or this column doesn't exist, there's an error
	category, err := extractGameCategoriesFromQueryResult(database, res)
	if err != nil {
		return nil, err
	}

	//Category is first from slice returned by extractGameCategoriesFromQueryResults
	return category[0], nil
}

// Get game category from ID
func GetGameCategoryByID(database *sql.DB, id int64) (*GameCategory, error) {
	//Query database for this game category
	cols := []string{
		db.ColGameCategoryName,
		db.ColGameCategoryEstimate,
		db.ColGameCategoryGameID,
		db.ColGameCategoryNumCollectibles,
		db.ColGameCategoryID,
	}
	table := db.TableGameCategories
	where := []db.WhereCondition{
		{ColName: db.ColGameCategoryID, Op: db.Equals, Value: id},
	}

	//Build and execute statement
	stmt := db.BuildSelectStatement(cols, table, where)
	res, err := db.ExecuteQueries(database, []db.SQLStatement{stmt})
	if err != nil {
		return nil, err
	}

	//Get actual categories from the results map
	category, err := extractGameCategoriesFromQueryResult(database, res)
	if err != nil {
		return nil, err
	}

	//Category is first element in the slice returned by extractGameCategoriesFromQueryResults
	return category[0], nil
}

// Checks if gamecategory exists
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

// Helper function for extracing game categories from SQL results
func extractGameCategoriesFromQueryResult(database *sql.DB, results map[string][]any) ([]*GameCategory, error) {
	//Get game category from the query. If there's none or this column doesn't exist, there's an error
	names, exists := results[db.ColGameCategoryName]

	//Check if category exists
	if !exists || len(names) == 0 {
		return nil, GameCategoryDoesNotExistErr
	}

	//Get fields from query results
	//Error handling could be a bit better here probably
	out := make([]*GameCategory, 0)
	for i := range names {
		cName, ok := results[db.ColGameCategoryName][i].(string)
		if !ok {
			return nil, errors.New("unknown error")
		}

		cEstimate, ok := results[db.ColGameCategoryEstimate][i].(string)
		if !ok {
			return nil, errors.New("unknown error")
		}

		cGameID, ok := results[db.ColGameCategoryGameID][i].(int64)
		if !ok {
			return nil, errors.New("unknown error")
		}

		cNumCollectibles, ok := results[db.ColGameCategoryNumCollectibles][i].(int64)
		if !ok {
			return nil, errors.New("unknown error")
		}

		cID, ok := results[db.ColGameCategoryID][i].(int64)
		if !ok {
			return nil, errors.New("unknown error")
		}

		//Get game from ID
		cGame, err := games.GetGameByID(database, cGameID)
		if err != nil {
			return nil, errors.New("unknown error: can't get category game")
		}

		//Add game category to output
		out = append(out, &GameCategory{
			Name:            repository.MakeNullableStr(cName),
			Estimate:        repository.MakeNullableStr(cEstimate),
			Game:            cGame,
			NumCollectibles: repository.MakeNullableInt(int(cNumCollectibles)),
			CategoryID:      cID,
		})
	}

	return out, nil
}
