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
	pTwitchID, err := twitch.Client.GetTwitchIDFromName(p.TwitchName.Value)
	if err != nil {
		return err
	}

	//Add values to the DB
	err = executeInsertStatements(database, p.Name.Value, pTwitchID, p.TwitchName.Value)
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
		newTwitchID, err := twitch.Client.GetTwitchIDFromName(newTwitchName.Value)
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
			db.ColSocialsPlatformUsername,
		}
		vals := []any {
			p.PlayerID,
			"twitch",
			newTwitchID,
			newTwitchName.Value,
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

//Queries DB for players and total result count
func QueryPlayers(database *sql.DB, playerQuery PlayerQuery, limit int, offset int) ([]*Player, int64, error) {
	//Build queries
	//Similar to game category, table needs to be specified to avoid ambiguity
	cols := []string {
		db.TablePlayers + "." + db.ColPlayerID,
		db.ColPlayerName,
		db.ColSocialsPlatformUserID,
	}
	table := db.JoinTables(db.TablePlayers, db.TableSocials, 
		db.GetOnClause(db.TablePlayers, db.TableSocials, db.ColPlayerID, db.ColSocialsPlayerID))

	twitchIDCache := make(map[string]string) //Maps ID to Name to avoid redundant Twitch API calls
	whereCons, err := getPlayerWhereCons(playerQuery, twitchIDCache)
	if err != nil {
		return nil, -1, err
	}
	order := db.Order{ColName: db.ColPlayerName, Direction: db.Ascending}

	//Execute query
	stmt := db.BuildSelectStatementWithLimitAndOffset(cols, table, whereCons, limit, offset, order)
	res, err := db.ExecuteQueries(database, []db.SQLStatement{stmt})
	if err != nil {
		return nil, -1, err
	}

	//Get count
	stmt = db.BuildCountStatement(cols, table, whereCons)
	count, err := db.ExecuteCountStatement(database, stmt)
	if err != nil {
		return nil, -1, err
	}

	//Output
	out := make([]*Player, 0)

	//If results are empty, return nothing
	if len(res[db.ColPlayerID]) == 0 {
		return out, count, nil
	}

	//Parse response
	out, err = parsePlayerQueryResponse(res, twitchIDCache)
	if err != nil {
		return nil, -1, err
	}

	return out, count, nil
}

//Gets player by name
func GetPlayerByName(database *sql.DB, name repository.NullableStr) (*Player, error) {
	if !name.Valid {
		return nil, repository.StringIsNullErr
	}

	//Get player from the query. If there's no names, check for twitch name as a fallback
	player, err := findPlayerByInternalName(database, name)
	if err != nil {
		return nil, err
	}
	names, exists := player[db.ColPlayerName]
	
	nameIsTwitchName := false
	//Search for player based on twitch name and update variables
	if len(names) == 0 {
		player, err = findPlayerByTwitchName(database, name)
		if err != nil {
			return nil, err
		}
		names, exists = player[db.ColPlayerName]
		
		//Twitch user was found, so this name is their twitch name
		if exists {
			nameIsTwitchName = true
		}
	}

	if !exists {
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
	var twitchName string
	if !nameIsTwitchName {
		col := []string{db.ColSocialsPlatformUserID}
		table := db.TableSocials
		where := []db.WhereCondition{{
			ColName: db.ColSocialsPlayerID,
			Op: db.Equals,
			Value: pID,
		}}
		stmt := db.BuildSelectStatement(col, table, where)
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
		twitchName, err = twitch.Client.GetTwitchNameFromID(twitchID)
		if err != nil {
			return nil, err
		}
	} else {twitchName = name.Value}

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
	id, err := twitch.Client.GetTwitchIDFromName(name.Value)
	if err != nil {
		return false, PlayerDoesNotExistErr
	}

	exists, err := db.RecordExists(database, db.TableSocials, db.ColSocialsPlatformUserID, id)
	if err != nil {
		return false, err
	}

	return exists, nil
}

//Helpers for finding player from DB
func findPlayerByInternalName(database *sql.DB, name repository.NullableStr) (map[string][]any, error){
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

	return player, nil
}

func findPlayerByTwitchName(database *sql.DB, name repository.NullableStr) (map[string][]any, error) {
	//Check if there's a player in this database with this Twitch name
	col := []string{db.ColPlayerName, db.TablePlayers + "." + db.ColPlayerID}
	on := db.GetOnClause(db.TablePlayers, db.TableSocials, db.ColPlayerID, db.ColSocialsPlayerID)
	table := db.JoinTables(db.TablePlayers, db.TableSocials, on)
	where := []db.WhereCondition{{
		ColName: db.TableSocials + "." + db.ColSocialsPlatformUsername,
		Op: db.Equals,
		Value: name.Value,
	}}

	stmt := db.BuildSelectStatement(col, table, where)
	player, err := db.ExecuteQueries(database, []db.SQLStatement{stmt})
	if err != nil {
		return nil, err
	}

	names := player[db.ColPlayerName]

	//Player still doesn't exist, now get twitch ID and search based on twitch ID
	if (len(names) == 0) {
		player, err = findPlayerByTwitchID(database, name)
		if err != nil {
			return nil, err
		}
	}

	return player, nil
}

func findPlayerByTwitchID(database *sql.DB, name repository.NullableStr) (map[string][]any, error) {
	id, err := twitch.Client.GetTwitchIDFromName(name.Value)
	if err != nil {
		//Error other than no user
		if err != twitch.UserCouldNotBeFoundErr {
			return nil, err
		}
	}

	//ID exists, check if they're in our DB
	col := []string{db.ColPlayerName, db.TablePlayers + "." + db.ColPlayerID}
	on := db.GetOnClause(db.TablePlayers, db.TableSocials, db.ColPlayerID, db.ColSocialsPlayerID)
	table := db.JoinTables(db.TablePlayers, db.TableSocials, on)
	where := []db.WhereCondition{{
		ColName: db.ColSocialsPlatformUserID,
		Op: db.Equals,
		Value: id,
	}}
	stmt := db.BuildSelectStatement(col, table, where)

	player, err := db.ExecuteQueries(database, []db.SQLStatement{stmt})
	if err != nil {
		return nil, err
	}

	//Update name and exists
	names := player[db.ColPlayerName]

	//If this player DOES exist, finally, update their twitch username
	if len(names) > 0 {
		col = []string{db.ColSocialsPlatformUsername}
		newVals := []any{name.Value}
		where = []db.WhereCondition{{
			ColName: db.ColSocialsPlatformUserID,
			Op: db.Equals,
			Value: id,
		}}
		stmt, err = db.BuildUpdateStatement(col, newVals, db.TableSocials, where)
		if err != nil {
			return nil, err
		}

		_, err := db.ExecuteStatements(database, []db.SQLStatement{stmt})
		if err != nil {
			return nil, err
		}
	}

	return player, nil
}

//Adds player and their twitch to DB atomically
func executeInsertStatements(database *sql.DB, playerName string, playerTwitchID string, playerTwitchName string) error {
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
		db.ColSocialsPlatformUsername,
	}
	table := db.TableSocials
	vals := []any {
		playerID,
		"twitch",
		playerTwitchID,
		playerTwitchName, 
	}
	socialsStmt := db.BuildInsertStatement(cols, table, vals)
	_, err = tx.Exec(socialsStmt.Stmt, socialsStmt.Args...)
	if err != nil {
		return err
	}

	return tx.Commit()
}