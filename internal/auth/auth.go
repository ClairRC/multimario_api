package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"

	"github.com/multimario_api/internal/db"
)

//Auth levels
type AuthLevel int

const (
	AuthNone AuthLevel = -1
	AuthVerified AuthLevel = 0
	AuthAdmin AuthLevel = 1
	AuthSuperAdmin AuthLevel = 2
)

func KeyMeetsLevel(database *sql.DB, key string, level AuthLevel) (bool, error) {
	//Get the key auth level from DB
	authLevel, err := getAPIKeyLevel(database, key)
	if err != nil {
		return false, err
	}

	return authLevel >= level, nil
}

//Generates random API key
func GetAPIKey(database *sql.DB, twitchID string) (string, error) {
	//Check if twitchID is already in db
	cols := []string{
		db.ColAPIKeysKey,
	}
	whereCon := []db.WhereCondition{{
		ColName: db.ColAPIKeyTwitchID,
		Op: db.Equals,
		Value: twitchID,
	}}

	//Get response
	stmt := db.BuildSelectStatement(cols, db.TableAPIKeys, whereCon)
	res, err := db.ExecuteQueries(database, []db.SQLStatement{stmt})
	if err != nil {
		return "", err
	}

	//If there's no key, gen a new one and add it, otherwise return the key
	if len(res[db.ColAPIKeysKey]) > 0 {
		key, ok := res[db.ColAPIKeysKey][0].(string)
		if !ok {
			return "", errors.New("unknown error: stored key can't be parsed as string")
		}
		return key, nil
	}

	bytes := make([]byte, 32)
	rand.Read(bytes)
	newKey := hex.EncodeToString(bytes)

	err = addAPIKey(database, newKey, twitchID)
	if err != nil {
		return "", err
	}
	
	return newKey, nil
}

//Adds API key to db
func addAPIKey(database *sql.DB, key string, twitchID string) error {
	cols := []string {
		db.ColAPIKeysKey,
		db.ColAPIKeyTwitchID,
	}
	vals := []any {key, twitchID}

	//Get and execute statement
	stmt := db.BuildInsertStatement(cols, db.TableAPIKeys, vals)
	_, err := db.ExecuteStatements(database, []db.SQLStatement{stmt})
	return err
}

//Helper to get API key level
func getAPIKeyLevel(database *sql.DB, key string) (AuthLevel, error) {
	//Get the key from DB
	col := []string{
		db.ColAPIKeysKey,
		db.ColAPIKeysAuthLevel,
	}
	table := db.TableAPIKeys
	whereCon := []db.WhereCondition{{
		ColName: db.ColAPIKeysKey,
		Op: db.Equals,
		Value: key,
	}}

	//Get statement and execute query
	stmt := db.BuildSelectStatement(col, table, whereCon)
	res, err := db.ExecuteQueries(database, []db.SQLStatement{stmt})
	if err != nil {
		return AuthNone, err
	}

	//Check if this API key exists
	if len(res[db.ColAPIKeysKey]) == 0 {
		return AuthNone, nil
	}

	//Check the "auth_level" col
	adminLevel, ok := res[db.ColAPIKeysAuthLevel][0].(int64)
	if !ok {
		return AuthNone, errors.New("unable to parse adming status from database")
	}

	return AuthLevel(int(adminLevel)), nil
}