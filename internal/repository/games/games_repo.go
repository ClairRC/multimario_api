package games

import (
	"database/sql"
	"errors"

	"github.com/multimario_api/internal/db"
	"github.com/multimario_api/internal/repository"
)

type Game struct {
	Name repository.NullableStr
	GameID int64
}

type GameQuery struct {
	Names []string
}

//Default error instantialtion
var GameDoesNotExistErr error = errors.New("game does not exist")

/*
* Game Constructor
*/

// Create new game instance
func NewGame(name repository.NullableStr) (*Game, error) {
	if !name.Valid {
		return nil, repository.StringIsNullErr
	}

	return &Game{Name: name}, nil
}

/*
* Game Methods
*/

//Add game
func (g *Game) Add(database *sql.DB) error {
	if g.GameID != 0 {
		return errors.New("game already exists")
	}

	//Build SQL statements
	stmt := db.BuildInsertStatement([]string{db.ColGameName}, db.TableGames, []any{g.Name.Value})
	
	ids, err := db.ExecuteStatements(database, []db.SQLStatement{stmt})
	if err != nil {
		return err
	}
	g.GameID = ids[0]

	return nil
}

//Update game
func (g *Game) Update(database *sql.DB, newName repository.NullableStr) error {
	//If game ID is 0, game doesn't exist
	if g.GameID == 0 {
		return GameDoesNotExistErr
	}

	stmts := make([]db.SQLStatement, 0, 1)

	//Get name update statement
	if newName.Valid {
		nameStmt, err := db.BuildUpdateStatement(
			[]string{db.ColGameName}, 
			[]any{newName.Value}, 
			db.TableGames, 
			[]db.WhereCondition{{ColName: db.ColGameID, Op: db.Equals, Value: g.GameID}})

		if err != nil {
			return err
		}

		stmts = append(stmts, nameStmt)
	}

	//No new updates, return nil
	if len(stmts) == 0 {
		return nil
	}

	_, err := db.ExecuteStatements(database, stmts)
	return err
}

/*
* Game Helpers
*/

//Gets games based on game query
func QueryGames(database *sql.DB, gameQuery GameQuery) ([]*Game, error) {
	out := make([]*Game, 0)

	//Get SQL query values
	cols := []string {
		db.ColGameID,
		db.ColGameName,
	}
	table := db.TableGames
	whereCons := getGameWhereCons(gameQuery)

	//Execute query
	stmt := db.BuildSelectStatement(cols, table, whereCons)
	res, err := db.ExecuteQueries(database, []db.SQLStatement{stmt})
	if err != nil {
		return nil, err
	}

	if len(res[db.ColGameID]) == 0 {
		return out, nil
	} //No results, return empty

	//Parse results and return them
	return parseQueryResponse(res), nil
}

//Gets game by name
func GetGameByName(database *sql.DB, name repository.NullableStr) (*Game, error) {
	if !name.Valid {
		return nil, repository.StringIsNullErr
	}

	//Query database for this game
	col := []string {db.ColGameName, db.ColGameID}
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
	ids, exists := game[db.ColGameID]
	if !exists || len(ids) == 0 {
		return nil, GameDoesNotExistErr
	}

	gName, ok := names[0].(string)
	if !ok {
		return nil, errors.New("unknown error: unable to parse game name")
	}
	gID, ok := ids[0].(int64)
	if !ok {
		return nil, errors.New("unknown error: unable to parse game id")
	}

	return &Game{Name: repository.MakeNullableStr(gName), GameID: gID}, nil
}

//Gets game by id
func GetGameByID(database *sql.DB, id int64) (*Game, error) {
	if id == 0 {
		return nil, GameDoesNotExistErr
	}

	//Query database for this game
	col := []string {db.ColGameName, db.ColGameID}
	table := db.TableGames
	where := []db.WhereCondition{
		{ColName: db.ColGameID, Op: db.Equals, Value: id},
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
	ids, exists := game[db.ColGameID]
	if !exists || len(ids) == 0 {
		return nil, GameDoesNotExistErr
	}

	gName, ok := names[0].(string)
	if !ok {
		return nil, errors.New("unknown error: unable to parse game name")
	}
	gID, ok := ids[0].(int64)
	if !ok {
		return nil, errors.New("unknown error: unable to parse game id")
	}

	return &Game{Name: repository.MakeNullableStr(gName), GameID: gID}, nil
}

//Checks if game already exists
func GameExistsByName(database *sql.DB, name repository.NullableStr) (bool, error) {
	if !name.Valid {
		return false, nil //Invalid game means it just doesn't exist
	}

	exists, err := db.RecordExists(database, db.TableGames, db.ColGameName, name.Value)
	if err != nil {
		return false, err
	}

	return exists, nil
}

//Gets game ID from the name
/*
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

//Gets game name from ID
func GetGameNameFromID(database *sql.DB, id int64) (string, error) {
	//Build SQL query
	stmt := db.BuildSelectStatement(
		[]string{db.ColGameName}, 
		db.TableGames, 
		[]db.WhereCondition{{ColName: db.ColGameID, Op: db.Equals, Value: id}},
	)

	res, err := db.ExecuteQueries(database, []db.SQLStatement{stmt})
	if err != nil {
		return "", err
	}

	//Get name from result
	if name, exists := res[db.ColGameName]; exists {
		//name exists but isnt string for some reason (shouldnt happen lowkey), otherwise return id
		if v, ok := name[0].(string); !ok {
			return "", errors.New("unexpected type for game name")
		} else {
			return v, nil
		}
	}

	//game doesn't exist
	return "", GameDoesNotExistErr
}
	*/