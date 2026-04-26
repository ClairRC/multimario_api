package games

import (
	"database/sql"
	"errors"

	"github.com/multimario_api/internal/db"
	"github.com/multimario_api/internal/repository"
)

type Game struct {
	Name repository.NullableStr
}


//Default error instantialtion
var GameDoesNotExistErr error = errors.New("game does not exist")

// Create new game instance
func NewGame(name repository.NullableStr) (*Game, error) {
	if !name.Valid {
		return nil, repository.StringIsNullErr
	}

	return &Game{Name: name}, nil
}

//Gets game by name
func GetGameByName(database *sql.DB, name repository.NullableStr) (*Game, error) {
	if !name.Valid {
		return nil, repository.StringIsNullErr
	}

	//Query database for this game
	col := []string {db.ColGameName}
	table := db.TableGames
	where := []db.WhereCondition{
		{ColName: db.ColGameName, Op: db.Equals, Value: name.Value},
	}

	stmt := db.BuildSelectStatement(col, table, where)
	game, err := db.ExecuteQueries(database, []db.SQLStatement{stmt})
	if err != nil {
		return nil, err
	}

	//Get games from the query. If there's none or this column doesn't exist, there's an error
	names, exists := game[db.ColGameName]
	if !exists || len(names) == 0 {
		return nil, GameDoesNotExistErr
	}

	gName, ok := game[db.ColGameName][0].(string)
	if !ok {
		return nil, errors.New("unknown error")
	}

	return &Game{Name: repository.MakeNullableStr(gName)}, nil
}

//Add game
func (g *Game) Add(database *sql.DB) error {
	//Build SQL statements
	stmt := db.BuildInsertStatement([]string{db.ColGameName}, db.TableGames, []any{g.Name})
	
	return db.ExecuteStatements(database, []db.SQLStatement{stmt})
}

//Update game
func (g *Game) Update(database *sql.DB, newName repository.NullableStr) error {
	stmts := make([]db.SQLStatement, 0, 2)

	//Get name update statement
	if newName.Valid {
		nameStmt, err := db.BuildUpdateStatement(
			[]string{db.ColGameName}, 
			[]any{newName.Value}, 
			db.TableGames, 
			[]db.WhereCondition{{ColName: db.ColGameName, Op: db.Equals, Value: g.Name.Value}})

		if err != nil {
			return err
		}

		stmts = append(stmts, nameStmt)
	}

	return db.ExecuteStatements(database, stmts)
}

//Checks if game already exists
func GameExistsByName(database *sql.DB, name repository.NullableStr) (bool, error) {
	if !name.Valid {
		return false, repository.StringIsNullErr
	}

	exists, err := db.RecordExists(database, db.TableGames, db.ColGameName, name.Value)
	if err != nil {
		return false, err
	}

	return exists, nil
}

//Helpers for querying DB
func GetGameIDFromName(database *sql.DB, name string) (int64, error){
	//Build SQL query
	stmt := db.BuildSelectStatement(
		[]string{db.ColGameID}, 
		db.TableGames, 
		[]db.WhereCondition{{ColName: db.ColGameName, Op: db.Equals, Value: name}},
	)

	res, err := db.ExecuteQueries(database, []db.SQLStatement{stmt})
	if err != nil {
		return -1, err
	}

	//Get ID from result
	if ids, exists := res[db.ColGameID]; exists {
		//ID exists but isnt int for some reason (shouldnt happen lowkey), otherwise return id
		if v, ok := ids[0].(int64); !ok {
			return -1, errors.New("unexpected type for game id")
		} else {
			return v, nil
		}
	}

	//ID doesn't exist
	return -1, GameDoesNotExistErr
}
