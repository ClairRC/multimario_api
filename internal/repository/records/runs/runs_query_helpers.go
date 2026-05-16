package runs

import (
	"database/sql"

	"github.com/multimario_api/internal/db"
	"github.com/multimario_api/internal/repository"
	"github.com/multimario_api/internal/repository/gamecategories"
)

//File for Runs query helpers

//Gets run where conditions
func getRunWhereCons(query RunQuery) []db.WhereCondition {
	out := make([]db.WhereCondition, 0)

	//Get runs belonging to each player
	playerWherePtr := repository.GetWhereCondition(db.ColPlayerName, query.PlayerNames, db.Equals)
	if playerWherePtr != nil {
		out = append(out, *playerWherePtr)
	}

	//Get runs of a certain category
	catWherePtr := repository.GetWhereCondition(db.ColGameCategoryName, query.GameCategories, db.Equals)
	if catWherePtr != nil {
		out = append(out, *catWherePtr)
	}

	//Get runs of a list of races
	raceWherePtr := repository.GetWhereCondition(db.ColRaceID, query.RaceIDs, db.Equals)
	if raceWherePtr != nil {
		out = append(out, *raceWherePtr)
	}
	
	
	return out
}

//Gets table for run queries
//Needs Players for player names, game categories for gamecat name, and records-races for race IDs
func getRunQueryTable() string {
	on := db.GetOnClause(db.TableRuns, db.TableGameCategories, db.ColRunGameCategoryID, db.ColGameCategoryID)
	table := db.JoinTables(db.TableRuns, db.TableGameCategories, on)

	on = db.GetOnClause(db.TableRuns, db.TableRecords, db.ColRunRaceRecordID, db.ColRecordID)
	table = db.JoinTables(table, db.TableRecords, on)

	on = db.GetOnClause(db.TableRecords, db.TableRaces, db.ColRecordsRaceID, db.ColRaceID)
	table = db.JoinTables(table, db.TableRaces, on)

	return table
}

//Parse query request
func parseRunQueryResults(database *sql.DB, res map[string][]any) ([]*Run, error){
	out := make([]*Run, 0)

	//Make runs for each result
	for i := range len(res[db.ColRunID]) {
		//Parse required values
		runID, ok := res[db.ColRunID][i].(int64)
		if !ok {
			continue
		}

		runEstimate, ok := res[db.ColRunEstimate][i].(string)
		if !ok {
			continue
		}

		//Parse optional values
		runTime := res[db.ColRunTime][i]

		//Get category from category ID
		catID, ok := res[db.ColRunGameCategoryID][i].(int64)
		if !ok {
			continue
		}
		cat, err := gamecategories.GetGameCategoryByID(database, catID)
		if err != nil {
			continue
		}

		//Get player name and race IDs specifically for the get response because stupid 
		playerName, ok := res[db.ColPlayerName][i].(string)
		if !ok {
			continue
		}
		raceID, ok := res[db.ColRaceID][i].(int64)
		if !ok {
			continue
		}

		newRun := &Run{
			Category: cat,
			Time: repository.MakeNullableStr(runTime),
			Estimate: repository.MakeNullableStr(runEstimate),
			RunID: runID,
			
			PlayerName: playerName,
			RaceID: raceID,
		}

		out = append(out, newRun)
	}

	return out, nil
}