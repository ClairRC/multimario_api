package races

import (
	"database/sql"
	"errors"

	"github.com/multimario_api/internal/db"
	"github.com/multimario_api/internal/repository"
	"github.com/multimario_api/internal/repository/racecategories"
)

type Race struct {
	Date repository.NullableStr
	StartTime repository.NullableStr
	Status repository.NullableStr
	RaceCategory *racecategories.RaceCategory
	RaceID int64 //DB ID for race. Defaults to 0
}

type RaceQuery struct {
	IDs []int64
	BeforeDates []string
	AfterDates []string
	OnDates []string
	Categories []string
	Statuses []string
}

//Errors
var RaceDoesNotExistErr error = errors.New("race does not exist")

/*
* Race Constructor
*/

// Create new race catgegory instance
func NewRace(database *sql.DB, date repository.NullableStr, startTime repository.NullableStr, status repository.NullableStr, categoryName repository.NullableStr) (*Race, error) {
	//Check required fields exist
	if !status.Valid {
		return nil, repository.StringIsNullErr
	}

	if !categoryName.Valid {
		return nil, repository.StringIsNullErr
	}

	if !status.Valid {
		return nil, repository.StringIsNullErr
	}

	//Get race category
	raceCategory, err := racecategories.GetRaceCategoryByName(database, categoryName)
	if err != nil {
		return nil, err
	}

	//Get Race output
	return &Race{
		Date: date,
		StartTime: startTime,
		Status: status,
		RaceCategory: raceCategory,
	}, nil
}

/*
* Race Methods
*/

// Add race. Returns race ID
func (r *Race) Add(database *sql.DB) error {
	if r.RaceID != 0 {
		return errors.New("race already exists")
	}

	//Get race category ID
	raceCatID := r.RaceCategory.CategoryID

	//If this is 0, the race category is invalid
	if raceCatID == 0 {
		return racecategories.RaceCategoryDoesNotExistErr
	}
	
	//Add race
	cols := []string {
		db.ColRaceRaceCategoryID,
		db.ColRaceDate,
		db.ColRaceStartTime,
		db.ColRaceStatus,
	}
	vals := []any {
		raceCatID,
		r.Date.NullableValue(),
		r.StartTime.NullableValue(),
		r.Status.Value,
	}

	add := db.BuildInsertStatement(cols, db.TableRaces, vals)
	
	//Execute insert and get id
	ids, err := db.ExecuteStatements(database, []db.SQLStatement{add})
	if err != nil {
		return err
	}

	//Make sure IDs isnt empty to avoid a panic
	if len(ids) == 0 {
		return errors.New("unknown error: no race id found")
	}

	//Get race ID and return
	r.RaceID = ids[0]
	return nil
}

// Update race
func (r *Race) Update(database *sql.DB, newDate repository.NullableStr, 
	newStartTime repository.NullableStr, newStatus repository.NullableStr) error {
		//TODO: Currently, this doesn't let you update a value to NULL

		//Check if race ID exists
		if r.RaceID == 0 {
			return RaceDoesNotExistErr
		}

		//Build Update statement parameters
		cols := make([]string, 0)
		vals := make([]any, 0)
		whereCon := []db.WhereCondition {{
			ColName: db.ColRaceID,
			Op: db.Equals,
			Value: r.RaceID,
		}}

		if newDate.Valid {
			cols = append(cols, db.ColRaceDate)
			vals = append(vals, newDate.Value)
		}
		if newStartTime.Valid {
			cols = append(cols, db.ColRaceStartTime)
			vals = append(vals, newStartTime.Value)
		}
		if newStatus.Valid {
			cols = append(cols, db.ColRaceStatus)
			vals = append(vals, newStatus.Value)
		}

		//If Cols/Vals is empty, no updates needed
		if len(cols) == 0 {
			return nil
		}

		update, err := db.BuildUpdateStatement(cols, vals, db.TableRaces, whereCon)
		if err != nil { return err }

		//Execute update
		_, err = db.ExecuteStatements(database, []db.SQLStatement{update})
		return err
}

/*
* Race Helpers
*/

//Query races
func QueryRaces(database *sql.DB, raceQuery RaceQuery) ([]*Race, error) {
	out := make([]*Race, 0)
	
	//Get query info
	cols := []string{
		db.ColRaceID,
		db.ColRaceDate,
		db.ColRaceStartTime,
		db.ColRaceStatus,
		db.ColRaceCategoryName,
	}
	table := db.JoinTables(db.TableRaces, db.TableRaceCategories, db.ColRaceRaceCategoryID, db.ColRaceCategoryID) //Races and Racecategories needed for information
	whereCons := getRaceWhereCons(raceQuery)

	//Execute query
	stmt := db.BuildSelectStatement(cols, table, whereCons)
	res, err := db.ExecuteQueries(database, []db.SQLStatement{stmt})
	if err != nil {
		return nil, err
	}

	if len(res[db.ColRaceID]) == 0 {
		return out, nil
	} //No results, return empty slice

	out, err = parseRaceQueryResponse(database, res)
	if err != nil {
		return nil, err
	}

	return out, nil
}

// Get race
func GetRaceByID(database *sql.DB, id int64) (*Race, error) {
	//Get select statement and execute it
	cols := []string{
		db.ColRaceDate,
		db.ColRaceStartTime,
		db.ColRaceStatus,
		db.ColRaceRaceCategoryID,
	}
	whereCon := []db.WhereCondition{{
		ColName: db.ColRaceID, 
		Op: db.Equals, 
		Value: id,
	}}

	raceStmt := db.BuildSelectStatement(cols, db.TableRaces, whereCon)
	raceRes, err := db.ExecuteQueries(database, []db.SQLStatement{raceStmt})
	if err != nil {
		return nil, err
	}

	//If there's no race category, that means this must not exist.
	if len(raceRes[db.ColRaceRaceCategoryID]) == 0 {
		return nil, RaceDoesNotExistErr
	}

	//Get race category name
	whereCon = []db.WhereCondition{{
		ColName: db.ColRaceCategoryID,
		Op: db.Equals,
		Value: raceRes[db.ColRaceRaceCategoryID][0],
	}}
	raceCatStmt := db.BuildSelectStatement([]string{db.ColRaceCategoryName}, db.TableRaceCategories, whereCon)
	raceCatRes, err := db.ExecuteQueries(database, []db.SQLStatement{raceCatStmt})
	if err != nil {
		return nil, err
	}

	//Get race category
	raceCatName := repository.MakeNullableStr(raceCatRes[db.ColRaceCategoryName][0])
	raceCat, err := racecategories.GetRaceCategoryByName(database, raceCatName)
	if err != nil {
		return nil, err
	}

	//Return Race
	return &Race{
		Date: repository.MakeNullableStr(raceRes[db.ColRaceDate][0]),
		StartTime: repository.MakeNullableStr(raceRes[db.ColRaceStartTime][0]),
		Status: repository.MakeNullableStr(raceRes[db.ColRaceStatus][0]),
		RaceCategory: raceCat,
		RaceID: id,
	}, nil
}