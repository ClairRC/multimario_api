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
func GetRunFromID(database *sql.DB, runID int64) (*Run, error) {
	if runID == 0 {
		return nil, RunDoesNotExistErr
	}

	//Get run values from DB
	cols := []string {
		db.ColRunGameCategoryID,
		db.ColRunTime,
		db.ColRunEstimate,
	}
	table := db.TableRuns
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