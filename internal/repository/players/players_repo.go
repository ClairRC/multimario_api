package players

import (
	"database/sql"
	"errors"

	"github.com/multimario_api/internal/db"
	"github.com/multimario_api/internal/repository"
)

type Player struct {
	Name repository.NullableStr
	TwitchName repository.NullableStr
	PlayerID int64 //DB id of player. Defaults to 0 for unadded player
}

//Default instantialtion
var PlayerDoesNotExistErr error = errors.New("player does not exist")

/*
* Player Constructor
*/

// Create new player instance
func NewPlayer(name repository.NullableStr, twitchName repository.NullableStr) (*Player, error) {
	if !name.Valid {
		return nil, repository.StringIsNullErr
	}

	if !twitchName.Valid {
		return nil, repository.StringIsNullErr
	}

	//TODO: Add check to make sure TwitchName exists and add a Socials row for this
	return &Player{Name: name, TwitchName: twitchName}, nil
}

/*
* Player Methods
*/

//Add player
func (p *Player) Add(database *sql.DB) error {
	if p.PlayerID != 0 {
		return errors.New("player already exists")
	}

	//TODO: Make sure to also get twitch ID to add to DB

	//Build SQL statements
	stmt := db.BuildInsertStatement([]string{db.ColPlayerName}, db.TablePlayers, []any{p.Name.Value})
	
	ids, err := db.ExecuteStatements(database, []db.SQLStatement{stmt})
	if err != nil {
		return err
	}
	p.PlayerID = ids[0] //Register ID to player

	return nil
}

//Update player
func (p *Player) Update(database *sql.DB, newName repository.NullableStr, newTwitchName repository.NullableStr) error {
	//If player ID is 0, player isn't in DB
	if p.PlayerID == 0 {
		return PlayerDoesNotExistErr
	}

	stmts := make([]db.SQLStatement, 0, 2)

	//If ID is valid, make the statement for that
	if newTwitchName.Valid {
		//TODO: Add logic for updating twitch ID for new twitch account
	}

	//Get name update statement
	if newName.Valid {
		nameStmt, err := db.BuildUpdateStatement(
			[]string{db.ColPlayerName}, 
			[]any{newName.Value}, 
			db.TablePlayers, 
			[]db.WhereCondition{{ColName: db.ColPlayerID, Op: db.Equals, Value: p.PlayerID}})

		if err != nil {
			return err
		}

		stmts = append(stmts, nameStmt)
	}

	//No statements to execute means we're good
	if len(stmts) == 0 {
		return nil
	}

	_, err := db.ExecuteStatements(database, stmts)
	return err
}

/*
* Player Helpers
*/

//Gets player by name
func GetPlayerByName(database *sql.DB, name repository.NullableStr) (*Player, error) {
	if !name.Valid {
		return nil, repository.StringIsNullErr
	}

	//Query database for this player
	col := []string {db.ColPlayerName, db.ColPlayerID}
	table := db.TablePlayers
	where := []db.WhereCondition{
		{ColName: db.ColPlayerName, Op: db.Equals, Value: name.Value},
	}

	stmt := db.BuildSelectStatement(col, table, where)
	player, err := db.ExecuteQueries(database, []db.SQLStatement{stmt})
	if err != nil {
		return nil, err
	}

	//Get player from the query. If there's none or this column doesn't exist, there's an error
	names, exists := player[db.ColPlayerName]
	if !exists || len(names) == 0 {
		return nil, PlayerDoesNotExistErr
	}

	pName, ok := player[db.ColPlayerName][0].(string)
	if !ok {
		return nil, errors.New("unknown error: unable to parse player name")
	}

	pID, ok := player[db.ColPlayerID][0].(int64)
	if !ok {
		return nil, errors.New("unknown error getting player: unable to parse player ID")
	}

	return &Player{Name: repository.MakeNullableStr(pName), PlayerID: pID}, nil
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