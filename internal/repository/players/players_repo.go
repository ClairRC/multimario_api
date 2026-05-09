package players

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/multimario_api/internal/db"
	"github.com/multimario_api/internal/repository"
	"github.com/multimario_api/internal/twitch"
)

type Player struct {
	Name repository.NullableStr
	TwitchName repository.NullableStr
	PlayerID int64 //DB id of player. Defaults to 0 for unadded player
}

type PlayerQuery struct {
	Names []string
	TwitchNames []string
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

	//Twitch logins are lowercase, so make this lowercase just in case
	twitchName.Value = strings.ToLower(twitchName.Value)

	return &Player{Name: name, TwitchName: twitchName}, nil
}

/*
* Player Methods
*/

//Add player
func (p *Player) Add(database *sql.DB) error {
	/*
	* Note: Simliarly to the records repository file, its important that a player gets added
	* and their twitch also gets added since currently it is a hard requirement for joining a race.
	* Because of this, to guarantee atomicity this function deals with raw SQL instead of using the DB abstractions.
	* This should probably be fixed at some point.
	*/
	if p.PlayerID != 0 {
		return errors.New("player already exists")
	}

	//Get twitch ID from twitch name
	pTwitchID, err := twitch.GetTwitchIDFromName(p.TwitchName.Value)
	if err != nil {
		return err
	}

	//Add values to the DB
	err = executeInsertStatements(database, p.Name.Value, pTwitchID)
	return err
}

//Update player
func (p *Player) Update(database *sql.DB, newName repository.NullableStr, newTwitchName repository.NullableStr) error {
	//If player ID is 0, player isn't in DB
	if p.PlayerID == 0 {
		return PlayerDoesNotExistErr
	}

	stmts := make([]db.SQLStatement, 0, 2)

	//If twitch name is valid, make the statement for that
	if newTwitchName.Valid {
		//Get twitch ID
		newTwitchID, err := twitch.GetTwitchIDFromName(newTwitchName.Value)
		if err != nil {
			return err
		}
		//Delete current social table and make a new one
		twitchDel := db.BuildDeleteStatement(db.TableSocials, []db.WhereCondition{{
			ColName: db.ColSocialsPlayerID,
			Op: db.Equals,
			Value: p.PlayerID,
		}})
		stmts = append(stmts, twitchDel)

		//Add new table
		cols := []string {
			db.ColSocialsPlayerID,
			db.ColSocialsPlatformName,
			db.ColSocialsPlatformUserID,
		}
		vals := []any {
			p.PlayerID,
			"twitch",
			newTwitchID,
		}
		twitchAdd := db.BuildInsertStatement(cols, db.TableSocials, vals)
		stmts = append(stmts, twitchAdd)
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

//Queries DB for players
func QueryPlayers(database *sql.DB, playerQuery PlayerQuery) ([]*Player, error) {
	//TODO: Refactor this. The logic I think is sound but its a mess lowkey
	
	//Handle results
	out := make([]*Player, 0) //Output

	//Build queries
	cols := []string {
		db.ColPlayerID,
		db.ColPlayerName,
		db.ColSocialsPlatformUserID,
	}
	table := db.JoinTables(db.TablePlayers, db.TableSocials, db.ColPlayerID, db.ColSocialsPlayerID)
	whereCons := make([]db.WhereCondition, 0)

	//Get name where conditions
	var nameWherePtr *db.WhereCondition
	for i, name := range playerQuery.Names {
		if i == 0 {
			nameWherePtr = &db.WhereCondition{
				ColName: db.ColPlayerName,
				Op: db.Equals,
				Value: name,
				Ors: make([]db.OrCondition, 0),
			}
		} else {
			nameWherePtr.Ors = append(nameWherePtr.Ors, db.OrCondition{
				Op: db.Equals,
				Value: name,
			})
		}
	}
	if nameWherePtr != nil {
		whereCons = append(whereCons, *nameWherePtr)
	}

	//Get twitch name where conditions
	
	//Get Twitch IDs from twitch names
	//Map of Twitch IDs to name for parsing response
	twitchIDs := make(map[string]string)

	var twitchIDWherePtr *db.WhereCondition
	for i, twitchName := range playerQuery.TwitchNames {
		//Get twitch ID from the name
		id, err := twitch.GetTwitchIDFromName(twitchName)
		if err != nil {
			return nil, err
		}
		twitchIDs[id] = twitchName //Link ID to name

		if i == 0 {
			twitchIDWherePtr = &db.WhereCondition{
				ColName: db.ColSocialsPlatformUserID,
				Op: db.Equals,
				Value: id,
				Ors: make([]db.OrCondition, 0),
			}
		} else {
			twitchIDWherePtr.Ors = append(twitchIDWherePtr.Ors, db.OrCondition{
				Op: db.Equals,
				Value: id,
			})
		}
	}
	if twitchIDWherePtr != nil {
		whereCons = append(whereCons, *twitchIDWherePtr)
	}

	stmt := db.BuildSelectStatement(cols, table, whereCons)
	res, err := db.ExecuteQueries(database, []db.SQLStatement{stmt})
	if err != nil {
		return nil, err
	}

	//If results are empty, return nothing
	if len(res[db.ColPlayerID]) == 0 {
		return out, nil
	}

	//Loop through results and create players
	for i := range len(res[db.ColPlayerID]) {
		name := repository.MakeNullableStr(res[db.ColPlayerName][i])
		twitchID, ok := res[db.ColSocialsPlatformUserID][i].(string) //Avoid panic for this type assertion
		if !ok {
			continue
		}

		twitchNameStr, cached := twitchIDs[twitchID]
		if !cached {
			alsoTwitchNameStr, err := twitch.GetTwitchNameFromID(twitchID)
			twitchNameStr = alsoTwitchNameStr //I'll fix this dont worry
			if err != nil {
				return nil, err
			}
		}
		twitchName := repository.MakeNullableStr(twitchNameStr)

		id := res[db.ColPlayerID][i].(int64)

		newPlayer := &Player {
			Name: name,
			TwitchName: twitchName,
			PlayerID: id,
		}
		out = append(out, newPlayer)
	}

	return out, nil
}

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

	//Get player twitch name
	col = []string{db.ColSocialsPlatformUserID}
	table = db.TableSocials
	where = []db.WhereCondition{{
		ColName: db.ColSocialsPlayerID,
		Op: db.Equals,
		Value: pID,
	}}
	stmt = db.BuildSelectStatement(col, table, where)
	socials, err := db.ExecuteQueries(database, []db.SQLStatement{stmt})
	if err != nil {
		return nil, err
	}

	if len(socials[db.ColSocialsPlatformUserID]) == 0 {
		return nil, errors.New("unknown error: player doesn't have a twitch registered")
	} //Avoids a panic in rare edge case

	twitchID, ok := socials[db.ColSocialsPlatformUserID][0].(string)
	if !ok {
		return nil, errors.New("unknown error getting twitch information: unable to parse twitch id")
	}

	//Get name from id
	twitchName, err := twitch.GetTwitchNameFromID(twitchID)
	if err != nil {
		return nil, err
	}

	return &Player{Name: repository.MakeNullableStr(pName), TwitchName: repository.MakeNullableStr(twitchName), PlayerID: pID}, nil
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

//Checks if twitch name is already in use
func TwitchInUseByName(database *sql.DB, name repository.NullableStr) (bool, error) {
	if !name.Valid {
		return false, repository.StringIsNullErr
	}
	
	//Get twitch ID
	id, err := twitch.GetTwitchIDFromName(name.Value)
	if err != nil {
		return false, err
	}

	exists, err := db.RecordExists(database, db.TableSocials, db.ColSocialsPlatformUserID, id)
	if err != nil {
		return false, err
	}

	return exists, nil
}

//Adds player and their twitch to DB atomically
func executeInsertStatements(database *sql.DB, playerName string, playerTwitchID string) error {
	//Build SQL statements
	playerStmt := db.BuildInsertStatement([]string{db.ColPlayerName}, db.TablePlayers, []any{playerName})

	tx, err := database.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	//Add player to DB
	res, err := tx.Exec(playerStmt.Stmt, playerStmt.Args...)
	if err != nil {
		return err
	}
	playerID, err := res.LastInsertId() //Get player ID
	if err != nil {
		return err
	}

	//Add social table for this player
	cols := []string {
		db.ColSocialsPlayerID,
		db.ColSocialsPlatformName,
		db.ColSocialsPlatformUserID,
	}
	table := db.TableSocials
	vals := []any {
		playerID,
		"twitch",
		playerTwitchID,
	}
	socialsStmt := db.BuildInsertStatement(cols, table, vals)
	_, err = tx.Exec(socialsStmt.Stmt, socialsStmt.Args...)
	if err != nil {
		return err
	}

	return tx.Commit()
}