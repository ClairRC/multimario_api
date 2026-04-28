package players

import (
	"database/sql"
	"errors"

	"github.com/multimario_api/internal/db"
	"github.com/multimario_api/internal/repository"
)

type Player struct {
	Name repository.NullableStr
	TwitchID repository.NullableInt
}

//Default instantialtion
var PlayerDoesNotExistErr error = errors.New("player does not exist")

// Create new player instance
func NewPlayer(name repository.NullableStr, twitchID repository.NullableInt) (*Player, error) {
	if !name.Valid {
		return nil, repository.StringIsNullErr
	}

	return &Player{Name: name, TwitchID: twitchID}, nil
}

//Gets player by name
func GetPlayerByName(database *sql.DB, name repository.NullableStr) (*Player, error) {
	if !name.Valid {
		return nil, repository.StringIsNullErr
	}

	//Query database for this player
	col := []string {db.ColPlayerName}
	table := db.TablePlayers
	where := []db.WhereCondition{
		{ColName: db.ColPlayerName, Op: db.Equals, Value: name.Value},
	}

	stmt := db.BuildSelectStatement(col, table, where)
	player, err := db.ExecuteQueries(database, []db.SQLStatement{stmt})
	if err != nil {
		return nil, err
	}

	//Get players from the query. If there's none or this column doesn't exist, there's an error
	names, exists := player[db.ColPlayerName]
	if !exists || len(names) == 0 {
		return nil, PlayerDoesNotExistErr
	}

	pName, ok := player[db.ColPlayerName][0].(string)
	if !ok {
		return nil, errors.New("unknown error")
	}

	//TODO: Implement twtich ID stuff
	return &Player{Name: repository.MakeNullableStr(pName), TwitchID: repository.NULLInt}, nil
}

//Add player
func (p *Player) Add(database *sql.DB) error {
	//Build SQL statements
	stmt := db.BuildInsertStatement([]string{db.ColPlayerName}, db.TablePlayers, []any{p.Name})
	
	_, err := db.ExecuteStatements(database, []db.SQLStatement{stmt})
	return err
}

//Update player
func (p *Player) Update(database *sql.DB, newName repository.NullableStr, newTwitchID repository.NullableInt) error {
	stmts := make([]db.SQLStatement, 0, 2)

	//If ID is valid, make the statement for that
	if newTwitchID.Valid {
		playerID, err := getPlayerIDFromName(database, p.Name.Value)
		if err != nil {
			return err
		}

		idStmt, err := db.BuildUpdateStatement(
			[]string{db.ColPlatformID}, 
			[]any{newTwitchID.Value}, 
			db.TableSocials, 
			[]db.WhereCondition{{ColName: db.ColPlayerID, Op: db.Equals, Value: playerID}})

		if err != nil {
			return err
		}

		stmts = append(stmts, idStmt)
	}

	//Get name update statement
	if newName.Valid {
		nameStmt, err := db.BuildUpdateStatement(
			[]string{db.ColPlayerName}, 
			[]any{newName.Value}, 
			db.TablePlayers, 
			[]db.WhereCondition{{ColName: db.ColPlayerName, Op: db.Equals, Value: p.Name.Value}})

		if err != nil {
			return err
		}

		stmts = append(stmts, nameStmt)
	}

	_, err := db.ExecuteStatements(database, stmts)
	return err
}

//Checks if player already exists
func PlayerExistsByName(database *sql.DB, name repository.NullableStr) (bool, error) {
	if !name.Valid {
		return false, repository.StringIsNullErr
	}

	exists, err := db.RecordExists(database, db.TablePlayers, db.ColPlayerName, name.Value)
	if err != nil {
		return false, err
	}

	return exists, nil
}

//Helpers for querying DB
func getPlayerIDFromName(database *sql.DB, name string) (int64, error){
	//Build SQL query
	stmt := db.BuildSelectStatement(
		[]string{db.ColPlayerID}, 
		db.TablePlayers, 
		[]db.WhereCondition{{ColName: db.ColPlayerName, Op: db.Equals, Value: name}},
	)

	res, err := db.ExecuteQueries(database, []db.SQLStatement{stmt})
	if err != nil {
		return -1, err
	}

	//Get ID from result
	if ids, exists := res[db.ColPlayerID]; exists {
		//ID exists but isnt int for some reason (shouldnt happen lowkey), otherwise return id
		if v, ok := ids[0].(int64); !ok {
			return -1, errors.New("unexpected type for player id")
		} else {
			return v, nil
		}
	}

	//ID doesn't exist
	return -1, PlayerDoesNotExistErr
}
