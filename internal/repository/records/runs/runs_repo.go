package runs

import (
	"database/sql"
	"errors"

	"github.com/multimario_api/internal/db"
	"github.com/multimario_api/internal/repository"
	"github.com/multimario_api/internal/repository/gamecategories"
)

type Run struct {
	Category *gamecategories.GameCategory
	Time repository.NullableStr
	Estimate repository.NullableStr
	RunID int64 //Run ID. Defaults to 0

	//These fields are only present when querying runs from a GET request because they are needed for the response.
	//It's a band-aid fix, but since records generally manage runs, and runs are queried separately from records,
	//the Run needs to know the player and race id which means it needs to know the record, but the record is what manages the run.
	//This is really the best solution I could think of for this problem without entirely restructuring how records/runs work.
	//But it is not a clean solution. It's messy internally but hopefully cleaner for actual end users.
	PlayerName string
	RaceID int64
}

type RunQuery struct {
	PlayerNames []string
	GameCategories []string
	RaceIDs []int64
}

//Default Errors
var RunDoesNotExistErr error = errors.New("run does not exist")

/*
* Run Constructors
*/

//Returns run pointer and error
func NewRun(database *sql.DB, catName repository.NullableStr, 
	runTime repository.NullableStr, estimate repository.NullableStr) (*Run, error) {
		//Get category
		if !catName.Valid {
			return nil, repository.StringIsNullErr
		}
		cat, err := gamecategories.GetGameCategoryByName(database, catName)
		if err != nil {
			return nil, err
		}

		//Get default estimate if null
		if !estimate.Valid {
			estimate = cat.Estimate
		}

		return &Run {
			Category: cat,
			Time: runTime,
			Estimate: estimate,
		}, nil
	}

//Creates run with default fields and returns pointer to it
func NewDefaultRun(database *sql.DB, catName repository.NullableStr) (*Run, error) {
	//Get category
	cat, err := gamecategories.GetGameCategoryByName(database, catName)
	if err != nil {
		return nil, err
	}

	//Get default estimate
	estimate := cat.Estimate

	//Return run
	return &Run{
		Category: cat,
		Time: repository.NULLStr,
		Estimate: estimate,
	}, nil
}

/*
* Run Methods
*/

//Updates run info
func (r *Run) Update(database *sql.DB, newTime repository.NullableStr, newEstimate repository.NullableStr, newRunNum repository.NullableInt) error {
	if r.RunID == 0 {
		return RunDoesNotExistErr
	}
	
	//Update values
	cols := make([]string, 0, 3)
	table := db.TableRuns
	vals := make([]any, 0, 3)
	whereCon := []db.WhereCondition{{
		ColName: db.ColRunID,
		Op: db.Equals,
		Value: r.RunID,
	}}

	//Check each field
	/*
	 * TODO: This currently means that you can't "unfinish" a run. Same issue is present in Record.Update
	 * This should be fixed to avoid incorrect information
	*/	
	if newTime.Valid {
		cols = append(cols, db.ColRunTime)
		vals = append(vals, newTime.Value)
	}

	if newEstimate.Valid {
		cols = append(cols, db.ColRunEstimate)
		vals = append(vals, newEstimate.Value)
	}

	if newRunNum.Valid {
		cols = append(cols, db.ColRunNum)
		vals = append(vals, newRunNum.Value)
	}

	//If cols is empty, no updates needed 
	if len(cols) == 0 {
		return nil
	}

	//Otherwise, make the update
	stmt, err := db.BuildUpdateStatement(cols, vals, table, whereCon)
	if err != nil {
		return err
	}

	_, err = db.ExecuteStatements(database, []db.SQLStatement{stmt})
	return err
}

/*
* Run Helpers
*/

//Query runs
func QueryRuns(database *sql.DB, runQuery RunQuery) ([]*Run, error) {
	out := make([]*Run, 0)

	//Build Query
	cols := []string{
		db.ColRunGameCategoryID, 
		db.ColRunTime,
		db.ColRunEstimate,
		db.ColRunID,
		db.ColRaceID,
		db.ColPlayerName,
	}
	table := getRunQueryTable()
	whereCons := getRunWhereCons(runQuery)

	//Execute query
	stmt := db.BuildSelectStatement(cols, table, whereCons)
	res, err := db.ExecuteQueries(database, []db.SQLStatement{stmt})
	if err != nil {
		return nil, err
	}

	//If results empty, return empty slice
	if len(res[db.ColRunID]) == 0 {
		return out, nil
	}

	out, err = parseRunQueryResults(database, res)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func GetRunFromID(database *sql.DB, runID int64) (*Run, error) {
	if runID == 0 {
		return nil, RunDoesNotExistErr
	}

	//Get run values from DB
	cols := []string {
		db.ColRunGameCategoryID,
		db.ColRunTime,
		db.ColRunEstimate,
		db.ColRecordID,
	}
	table := db.JoinTables(db.TableRuns, db.TableRecords, db.ColRunRaceRecordID, db.ColRecordID) //Join tables to get record ID
	table = db.JoinTables(table, db.TablePlayers, db.ColRecordsPlayerID, db.ColPlayerID)
	whereCon := []db.WhereCondition{{
		ColName: db.ColRunID,
		Op: db.Equals,
		Value: runID,
	}}

	stmt := db.BuildSelectStatement(cols, table, whereCon)
	res, err := db.ExecuteQueries(database, []db.SQLStatement{stmt})
	if err != nil {
		return nil, err
	}

	//Make sure run exists
	if len(res[db.ColRunGameCategoryID]) == 0 {
		return nil, RunDoesNotExistErr
	}

	//Get fields from result
	catID, ok := res[db.ColRunGameCategoryID][0].(int64)
	if !ok {
		return nil, errors.New("unknown error: unable to parse category id as int")
	}

	//Get game category from ID
	cat, err := gamecategories.GetGameCategoryByID(database, catID)
	if err != nil {
		return nil, err
	}

	//Return run
	return &Run {
		Category: cat,
		Time: repository.MakeNullableStr(res[db.ColRunTime][0]),
		Estimate: repository.MakeNullableStr(res[db.ColRunEstimate][0]),
		RunID: runID,
	}, nil
}

//Gets run from run record ID and category name
func GetRunFromRecordID(database *sql.DB, recordID int64, categoryName repository.NullableStr) (*Run, error){
	//Build query
	if recordID == 0 {
		return nil, RunDoesNotExistErr
	}
	if !categoryName.Valid {
		return nil, gamecategories.GameCategoryDoesNotExistErr
	}

	cols := []string {
		db.ColRunID,
		db.ColRunGameCategoryID,
		db.ColRunTime,
		db.ColRunEstimate,
		db.ColPlayerName,
	}

	//Join tables to search with category name
	table := db.JoinTables(db.TableRuns, db.TableGameCategories, db.ColRunGameCategoryID, db.ColGameCategoryID)
	table = db.JoinTables(table, db.TableRecords, db.ColRunRaceRecordID, db.ColRecordID)
	table = db.JoinTables(table, db.TablePlayers, db.ColRecordsPlayerID, db.ColPlayerID)
	whereCon := []db.WhereCondition{{
		ColName: db.ColGameCategoryName,
		Op: db.Equals,
		Value: categoryName.Value,
	}}

	//Execute statement
	stmt := db.BuildSelectStatement(cols, table, whereCon)
	res, err := db.ExecuteQueries(database, []db.SQLStatement{stmt})
	if err != nil {
		return nil, err
	}

	//Make sure results are not empty
	if len(res[db.ColRunID]) == 0 {
		return nil, RunDoesNotExistErr
	}
	runID, ok := res[db.ColRunID][0].(int64)
	if !ok || runID == 0 {
		return nil, RunDoesNotExistErr
	}

	//Get game category
	gameCat, err := gamecategories.GetGameCategoryByName(database, categoryName)
	if err != nil {
		return nil, err
	}

	//Return run
	return &Run {
		Category: gameCat,
		Time: repository.MakeNullableStr(res[db.ColRunTime][0]),
		Estimate: repository.MakeNullableStr(res[db.ColRunEstimate][0]),
		RunID: runID,
	}, nil
}