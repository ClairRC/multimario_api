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
	GameName repository.NullableStr
}

// Errors
var GameCategoryDoesNotExistErr error = errors.New("game category does not exist")

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
		Name:            name,
		Estimate:        estimate,
		NumCollectibles: numCollectibles,
		GameName:        gameName,
	}, nil
}

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

// Add game category
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

	_, err = db.ExecuteStatements(database, []db.SQLStatement{stmt})
	return err
}

// Update game category
func (c *GameCategory) Update(
	database *sql.DB, newName repository.NullableStr,
	newEstimate repository.NullableStr, newNumCollectibles repository.NullableInt,
	newGameName repository.NullableStr,
) error {

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

	if newGameName.Valid {
		//Get game ID from name
		gameID, err := games.GetGameIDFromName(database, newGameName.Value)
		if err != nil {
			return err
		} //Return if there's an error getting game ID

		cols = append(cols, db.ColGameCategoryGameID)
		newVals = append(newVals, gameID)
	}

	//If there's nothing new to update, just return
	if len(cols) == 0 {
		return nil
	}

	//Otherwise, build and execute statement
	whereCon := db.WhereCondition{
		ColName: db.ColGameCategoryName,
		Op:      db.Equals,
		Value:   c.Name.Value,
	}

	stmt, err := db.BuildUpdateStatement(cols, newVals, db.TableGameCategories, []db.WhereCondition{whereCon})
	if err != nil {
		return err
	} //Return error if can't build statement

	_, err = db.ExecuteStatements(database, []db.SQLStatement{stmt})
	return err
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

// Helper to get GameCategory ID from name
func GetGameCategoryIDFromName(database *sql.DB, name string) (int64, error) {
	whereCon := []db.WhereCondition{{
		ColName: db.ColGameCategoryName,
		Op:      db.Equals,
		Value:   name,
	}}

	stmt := db.BuildSelectStatement([]string{db.ColGameCategoryID}, db.TableGameCategories, whereCon)
	res, err := db.ExecuteQueries(database, []db.SQLStatement{stmt})
	if err != nil {
		return -1, err
	}

	if len(res[db.ColGameCategoryID]) == 0 {
		return -1, GameCategoryDoesNotExistErr
	}

	id, ok := res[db.ColGameCategoryID][0].(int64)
	if !ok {
		return -1, errors.New("unknown error: game category id isn't an int")
	}

	return id, nil
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

		//Get game name from ID
		cGameName, err := games.GetGameNameFromID(database, cGameID)
		if err != nil {
			return nil, errors.New("unknown error")
		}

		//Add game category to output
		out = append(out, &GameCategory{
			Name:            repository.MakeNullableStr(cName),
			Estimate:        repository.MakeNullableStr(cEstimate),
			GameName:        repository.MakeNullableStr(cGameName),
			NumCollectibles: repository.MakeNullableInt(int(cNumCollectibles)),
		})
	}

	return out, nil
}
