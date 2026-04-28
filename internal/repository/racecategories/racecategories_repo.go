package racecategories

import (
	"database/sql"
	"errors"

	"github.com/multimario_api/internal/db"
	"github.com/multimario_api/internal/repository"
	"github.com/multimario_api/internal/repository/gamecategories"
)

// Race category struct
type RaceCategory struct {
	Name           repository.NullableStr
	GameCategories []*gamecategories.GameCategory
}

// Default Erros
var RaceCategoryDoesNotExistErr = errors.New("race category does not exist")

// Create new race catgegory instance
func NewRaceCategory(database *sql.DB, name repository.NullableStr, gameCats []*gamecategories.GameCategory) (*RaceCategory, error) {
	if !name.Valid {
		return nil, repository.StringIsNullErr
	}

	//Check each game category
	//I'm stubbing this out because realistically this should be done by the handler,
	//and since each one is an extra database call it's just not necessary with the current setup.
	/*
		for _, v := range gameCats {
			//Throw an error if any of these don't exist
			if exists, err := gamecategories.GameCategoryExistsByName(database, v.Name) {
				return nil, repository.StringIsNullErr
			}
		}*/

	return &RaceCategory{Name: name, GameCategories: gameCats}, nil
}

// Get race category
func GetRaceCategoryByName(database *sql.DB, name repository.NullableStr) (*RaceCategory, error) {
	//Check name is valid
	if !name.Valid {
		return nil, repository.StringIsNullErr
	}

	//Get race category ID
	
	//TODO:The big issue here is that if this succeeds and the queries below fail there will be an orphaned race category
	//Not ideal, but okay for now
	raceCatID, err := GetRaceCategoryIDFromName(database, name.Value)
	if err != nil {
		return nil, err
	}

	//Get game categories that are part of this race category
	//These two steps can be turned into 1 with JOINs
	selectLinkingStmt := db.BuildSelectStatement(
		[]string{db.ColRaceCatGameCatGameCatgeoryID},
		db.TableRaceCatGameCat,
		[]db.WhereCondition{{
			ColName: db.ColRaceCatGameCatRaceCategoryID,
			Op:      db.Equals,
			Value:   raceCatID,
		}},
	)
	gameCatRes, err := db.ExecuteQueries(database, []db.SQLStatement{selectLinkingStmt})
	if err != nil {
		return nil, err
	}

	//For each game category ID, get the game category and add it to the game category slice
	gameCategories := make([]*gamecategories.GameCategory, 0)
	for _, v := range gameCatRes[db.ColRaceCatGameCatGameCatgeoryID] {
		//Get game category ID from result map
		gameCatID, ok := v.(int64)
		if !ok {
			return nil, errors.New("unexpected type for game category id")
		}

		//Get the game category and add it to slice
		newGameCat, err := gamecategories.GetGameCategoryByID(database, gameCatID)
		if err != nil {
			return nil, err
		}
		gameCategories = append(gameCategories, newGameCat)
	}

	//No errors thank god
	return &RaceCategory{Name: name, GameCategories: gameCategories}, nil
}

// Add race category
func (c *RaceCategory) Add(database *sql.DB) error {
	//Add race category to DB
	add := db.BuildInsertStatement([]string{db.ColRaceCategoryName}, db.TableRaceCategories, []any{c.Name.Value})
	_, err := db.ExecuteStatements(database, []db.SQLStatement{add})
	if err != nil {
		return err
	}

	//Get race category ID
	//NOTE This call isn't necessary. Execute statements should probably return the ID
	raceCatID, err := GetRaceCategoryIDFromName(database, c.Name.Value)
	if err != nil {
		return err
	}

	//Get statements for adding each linking table entry
	stmts := make([]db.SQLStatement, 0)

	//For each game category, make a statement for adding the linking entry
	for _, v := range c.GameCategories {
		gameCatID, err := gamecategories.GetGameCategoryIDFromName(database, v.Name.Value)
		if err != nil {
			return err
		}

		cols := []string{db.ColRaceCatGameCatRaceCategoryID, db.ColRaceCatGameCatGameCatgeoryID}
		vals := []any{raceCatID, gameCatID}

		newStmt := db.BuildInsertStatement(cols, db.TableRaceCatGameCat, vals)
		stmts = append(stmts, newStmt)
	}

	//Execute statements atomically
	_, err = db.ExecuteStatements(database, stmts)
	if err != nil {
		return err
	}

	//Return nil if execute succeeds
	return nil
}

// Update race category
func (c *RaceCategory) Update(database *sql.DB, newName repository.NullableStr, newGameCategories []*gamecategories.GameCategory) error {
	//Statements
	stmts := make([]db.SQLStatement, 0)
	raceCatID, err := GetRaceCategoryIDFromName(database, c.Name.Value)
	if err != nil {
		return err
	}

	//If new game categories isn't empty, then delete all the existing ones and replace them
	if len(newGameCategories) > 0 {
		deleteCatsStmt := db.BuildDeleteStatement(
			db.TableRaceCatGameCat,
			[]db.WhereCondition{{
				ColName: db.ColRaceCatGameCatRaceCategoryID,
				Op:      db.Equals,
				Value:   raceCatID},
			},
		)
		stmts = append(stmts, deleteCatsStmt)

		//Get statements for adding new categories to linking table
		for _, v := range newGameCategories {
			gameCatID, err := gamecategories.GetGameCategoryIDFromName(database, v.Name.Value)
			if err != nil {
				return err
			}

			cols := []string{db.ColRaceCatGameCatRaceCategoryID, db.ColRaceCatGameCatGameCatgeoryID}
			vals := []any{raceCatID, gameCatID}

			newStmt := db.BuildInsertStatement(cols, db.TableRaceCatGameCat, vals)
			stmts = append(stmts, newStmt)
		}
	}

	//Statement for updating name
	if newName.Valid {
		cols := []string{db.ColRaceCategoryName}
		newVals := []any{newName.Value}
		whereCon := []db.WhereCondition{{
			ColName: db.ColRaceCategoryID,
			Op:      db.Equals,
			Value:   raceCatID,
		}}

		updateStmt, err := db.BuildUpdateStatement(cols, newVals, db.TableRaceCategories, whereCon)
		if err != nil {
			return err
		}

		stmts = append(stmts, updateStmt)
	}

	//Atomically execute statements
	_, err = db.ExecuteStatements(database, stmts)
	if err != nil {
		return err
	}

	//Statements executed successfully
	return nil
}

// Checks if race category exists
func RaceCategoryExistsByName(database *sql.DB, name repository.NullableStr) (bool, error) {
	if !name.Valid {
		return false, repository.StringIsNullErr
	}

	exists, err := db.RecordExists(database, db.TableRaceCategories, db.ColRaceCategoryName, name.Value)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// Helper to get race category ID from the name
func GetRaceCategoryIDFromName(database *sql.DB, name string) (int64, error) {
	whereCon := []db.WhereCondition{{
		ColName: db.ColRaceCategoryName,
		Op:      db.Equals,
		Value:   name,
	}}

	stmt := db.BuildSelectStatement([]string{db.ColRaceCategoryID}, db.TableRaceCategories, whereCon)
	res, err := db.ExecuteQueries(database, []db.SQLStatement{stmt})
	if err != nil {
		return -1, err
	}

	if len(res[db.ColRaceCategoryID]) == 0 {
		return -1, RaceCategoryDoesNotExistErr
	}

	id, ok := res[db.ColRaceCategoryID][0].(int64)
	if !ok {
		return -1, errors.New("unknown error: race category id isn't an int")
	}

	return id, nil
}
