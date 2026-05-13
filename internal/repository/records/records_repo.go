package records

import (
	"database/sql"
	"errors"

	"github.com/multimario_api/internal/db"
	"github.com/multimario_api/internal/repository"
	"github.com/multimario_api/internal/repository/players"
	"github.com/multimario_api/internal/repository/races"
	"github.com/multimario_api/internal/repository/records/runs"
)

type Record struct {
	Player *players.Player
	Race *races.Race
	FinishTime repository.NullableStr
	NumCollected repository.NullableInt
	RecordID int64 //Record ID. Defaults to 0
}

type RecordQuery struct {
	PlayerNames []string
	RaceIDs []int64
	Categories []string
	BeforeDates []string
	AfterDates []string
	OnDates []string
	LowerThan []string
	HigherThan []string
}

//Default errors
var RecordDoesNotExistErr error = errors.New("race record does not exist")

/*
* Record Constructor
*/

//Creates new Record and returns a pointer to it
func NewRecord(database *sql.DB, raceID repository.NullableInt, 
	playerName repository.NullableStr, finishTime repository.NullableStr, numCollected repository.NullableInt) (*Record, error) {
		//Make sure required fields exist
		if !raceID.Valid {
			return nil, errors.New("race is invalid")
		}
		if !playerName.Valid {
			return nil, errors.New("player name is invalid")
		}

		//Default numCollected is 0. Should be handled by the handler, but I'll put it here for now too
		if !numCollected.Valid {
			numCollected = repository.MakeNullableInt(0)
		}

		//Get race and player
		player, err := players.GetPlayerByName(database, playerName)
		if err != nil {
			return nil, err
		}

		race, err := races.GetRaceByID(database, int64(raceID.Value))
		if err != nil {
			return nil, err
		}

		//Created, return new record
		return &Record{Race: race, Player: player, FinishTime: finishTime, NumCollected: numCollected}, nil
	}

/*
* Records Methods
*/

//Adds race to DB
func (r *Record) Add(database *sql.DB, runs []*runs.Run) error {
	/*
	* This function, unlike all my other repository layer functions, handles raw SQL.
	* This is intentional because the current database abstraction won't let me add a new Record
	* AND add each of the individual runs atomically. For this in particular, it is important
	* that these happen atomically. Since the Record is essentially a race signup, if a record
	* exists but not any runs, then it could cause complications with updating during the race.
	* The safest and easiest way to handle this is to guarantee atomicity in this function. 
	*/

	/*
	* Additionally, unlike most repository layer abstractions, Record is responsible for adding Runs.
	* This is because a run should not under any circumstance exist without a record.
	*/
	
	if r.RecordID != 0 {
		return errors.New("record already exists")
	}

	//Get player ID
	playerID := r.Player.PlayerID
	if playerID == 0 {
		return players.PlayerDoesNotExistErr
	}

	//Get race ID
	raceID := r.Race.RaceID
	if raceID == 0 {
		return races.RaceDoesNotExistErr
	}

	//Check that each run is not already part of a different record
	for _, run := range runs {
		if run.RunID != 0 {
			return errors.New("run is already in database")
		}
	}

	err := executeRecordInsertStatement(database, r, runs)

	return err
}

//Updates record
func (r *Record) Update(database *sql.DB, newFinishTime repository.NullableStr, newNumCollected repository.NullableInt) error {
	//Make sure record exists
	if r.RecordID == 0 {
		return RecordDoesNotExistErr
	}

	//Get update statement
	cols := make([]string, 0)
	vals := make([]any, 0)

	whereCon := []db.WhereCondition{{
		ColName: db.ColRecordID,
		Op: db.Equals,
		Value: r.RecordID,
	}}

	if newFinishTime.Valid {
		cols = append(cols, db.ColRecordsFinishTime)
		vals = append(vals, newFinishTime.Value)
	}

	if newNumCollected.Valid {
		cols = append(cols, db.ColRecordsNumCollected)
		vals = append(vals, newNumCollected.Value)
	}

	//If cols/vals is empty, no updates necessary
	if len(cols) == 0 {
		return nil
	}

	//Get statement and execute it
	updateStmt, err := db.BuildUpdateStatement(cols, vals, db.TableRecords, whereCon)
	if err != nil { return err }

	_, err = db.ExecuteStatements(database, []db.SQLStatement{updateStmt})

	return err
}

/*
* Records Helpers
*/

//Queries records. Returns slice of records or error
func QueryRecord(database *sql.DB, recordQuery RecordQuery) ([]*Record, error) {
	//Get record information first and then query for the runs. It can probably be done in 1 SQL statement, but my abstractions 
	//aren't really good enough for that. So this will work for this scale
	cols := []string {
		db.ColPlayerName,
		db.ColRecordsRaceID,
		db.ColRecordsFinishTime,
		db.ColRecordsNumCollected,
		db.ColRecordID,
	}
	table := getRecordQueryTable()
	whereCons := getRecordsWhereCons(recordQuery)

	//Execute queries
	stmt := db.BuildSelectStatement(cols, table, whereCons)
	res, err := db.ExecuteQueries(database, []db.SQLStatement{stmt})
	if err != nil {
		return nil, err
	}

	//Parse results
	out := make([]*Record, 0)
	if len(res[db.ColRecordID]) == 0 {
		return out, nil
	}

	out = parseRecordQuery(database, res)

	return out, nil
}

//Gets record from race ID and player name
func GetRecord(database *sql.DB, raceID repository.NullableInt, playerName repository.NullableStr) (*Record, error) {
	//Make sure input values are valid
	if !raceID.Valid || raceID.Value == 0 {
		return nil, races.RaceDoesNotExistErr
	}
	if !playerName.Valid {
		return nil, players.PlayerDoesNotExistErr
	}

	//Get player and race
	race, err := races.GetRaceByID(database, int64(raceID.Value))
	if err != nil { return nil, err }
	
	player, err := players.GetPlayerByName(database, playerName)
	if err != nil { return nil, err }

	//Get race values from DB
	cols := []string {
		db.ColRecordsFinishTime,
		db.ColRecordsNumCollected,
		db.ColRecordID,
	}
	whereCons := []db.WhereCondition{{
		ColName: db.ColRecordsPlayerID,
		Op: db.Equals,
		Value: player.PlayerID,
	},
	{
		ColName: db.ColRecordsRaceID,
		Op: db.Equals,
		Value: race.RaceID,
	}}

	//Get record
	stmt := db.BuildSelectStatement(cols, db.TableRecords, whereCons)
	res, err := db.ExecuteQueries(database, []db.SQLStatement{stmt})
	if err != nil { return nil, err }

	//Make sure record exists and its values be parsed as int
	if len(res[db.ColRecordID]) == 0 {
		return nil, RecordDoesNotExistErr
	}

	recordID, ok := res[db.ColRecordID][0].(int64)
	if !ok {
		return nil, errors.New("unknown error: record id can't be parsed as int")
	}

	finishTimeStr, ok := res[db.ColRecordsFinishTime][0].(string)
	var recordFinishTime repository.NullableStr
	if !ok {
		recordFinishTime = repository.NULLStr
	} else {
		recordFinishTime = repository.MakeNullableStr(finishTimeStr)
	}

	numCollected, ok := res[db.ColRecordsNumCollected][0].(int64)
	if !ok {
		return nil, errors.New("unknown error: number collected in record can't be parsed as int")
	}

	//Make record and return it
	out := &Record {
		Player: player,
		Race: race,
		FinishTime: recordFinishTime,
		NumCollected: repository.MakeNullableInt(numCollected),
		RecordID: recordID,
	}

	return out, nil
}

//Verify record exists
func RecordExists(database *sql.DB, recordID int64) (bool, error) {
	exists, err := db.RecordExists(database, db.TableRecords, db.ColRecordID, recordID)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func getRunsFromRecordID(database *sql.DB, recordID int64) ([]*runs.Run, error) {
	//Build select statement
	cols := []string {db.ColRunID}
	whereCon := []db.WhereCondition{{
		ColName: db.ColRunRaceRecordID,
		Op: db.Equals, 
		Value: recordID,
	}}

	stmt := db.BuildSelectStatement(cols, db.TableRuns, whereCon)
	runsRes, err := db.ExecuteQueries(database, []db.SQLStatement{stmt})
	if err != nil { return nil, err }

	//Make sure there are runs
	if len(runsRes[db.ColRunID]) == 0 {
		return nil, RecordDoesNotExistErr
	}

	//For each run, get the run and add it to the slice
	out := make([]*runs.Run, 0)
	for _, v := range runsRes[db.ColRunID] {
		//Make sure ID is int
		id, ok := v.(int64)
		if !ok {
			return nil, errors.New("unknown error: run id can't be parsed as int")
		}

		run, err := runs.GetRunFromID(database, id)
		if err != nil { return nil, err }

		out = append(out, run)
	}

	return out, nil
}

func executeRecordInsertStatement(database *sql.DB, record *Record, runs []*runs.Run) error {
	//Get values for adding new Record
	cols := []string {
		db.ColRecordsRaceID,
		db.ColRecordsPlayerID,
		db.ColRecordsNumCollected,
	}
	vals := []any {record.Race.RaceID, record.Player.PlayerID, record.NumCollected.Value}

	//Also add finish time if that is a valid value
	if record.FinishTime.Valid {
		cols = append(cols, db.ColRecordsFinishTime)
		vals = append(vals, record.FinishTime.Value)
	}
	
	recordStmt := db.BuildInsertStatement(cols, db.TableRecords, vals)

	//Begin transaction

	//Add record
	tx, err := database.Begin()
	if err != nil { return err }

	defer tx.Rollback()

	res, err := tx.Exec(recordStmt.Stmt, recordStmt.Args...)
	if err != nil { return err }

	recordID, err := res.LastInsertId()
	if err != nil { return err }

	//Add runs
	for i, run := range runs {
		cols = []string {
			db.ColRunRaceRecordID,
			db.ColRunGameCategoryID,
			db.ColRunEstimate,
			db.ColRunNum,
		}
		vals = []any{recordID, run.Category.CategoryID, run.Estimate.Value, i+1}

		//If time is not NULL, pass in time as well
		if run.Time.Valid {
			cols = append(cols, db.ColRunTime)
			vals = append(vals, run.Time.Value)
		}

		//Get and exectute statement
		runStmt := db.BuildInsertStatement(cols, db.TableRuns, vals)
		_, err := tx.Exec(runStmt.Stmt, runStmt.Args...)
		if err != nil { return err }
	} 

	err = tx.Commit()
	if err != nil {
		return err
	}

	record.RecordID = recordID
	return nil
}