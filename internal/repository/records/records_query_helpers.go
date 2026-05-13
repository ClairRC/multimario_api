package records

import (
	"database/sql"

	"github.com/multimario_api/internal/db"
	"github.com/multimario_api/internal/repository"
	"github.com/multimario_api/internal/repository/players"
	"github.com/multimario_api/internal/repository/races"
)

//Helpers for querying race records

//Get where conditions from query
func getRecordsWhereCons(query RecordQuery) []db.WhereCondition {
	out := make([]db.WhereCondition, 0)

	//Get where based on players
	playersWhereConPtr := repository.GetWhereCondition(db.ColPlayerName, query.PlayerNames, db.Equals)
	if playersWhereConPtr != nil {
		out = append(out, *playersWhereConPtr)
	}

	//Get where based on race IDs
	raceIDWhereConPtr := repository.GetWhereCondition(db.ColRecordsRaceID, query.RaceIDs, db.Equals)
	if raceIDWhereConPtr != nil {
		out = append(out, *raceIDWhereConPtr)
	}

	//Get where based on categories
	categoryNameWhereConPtr := repository.GetWhereCondition(db.ColRaceCategoryName, query.Categories, db.Equals)
	if categoryNameWhereConPtr != nil {
		out = append(out, *categoryNameWhereConPtr)
	}

	//Get where based on before dates
	beforeWhereConPtr := repository.GetWhereCondition(db.ColRaceDate, query.BeforeDates, db.LessThan)
	if beforeWhereConPtr != nil {
		out = append(out, *beforeWhereConPtr)
	}

	//Get where based on after dates
	afterWhereConPtr := repository.GetWhereCondition(db.ColRaceDate, query.AfterDates, db.GreaterThan)
	if afterWhereConPtr != nil {
		out = append(out, *afterWhereConPtr)
	}

	//Get where based on on dates
	onWhereConPtr := repository.GetWhereCondition(db.ColRaceDate, query.OnDates, db.Equals)
	if onWhereConPtr != nil {
		out = append(out, *onWhereConPtr)
	}

	//Get runs based on higher than times
	higherThanWhereConPtr := repository.GetWhereCondition(db.ColRecordsFinishTime, query.HigherThan, db.GreaterThan)
	if higherThanWhereConPtr != nil {
		out = append(out, *higherThanWhereConPtr)
	}

	//Get runs based on lower than times
	lowerThanWhereConPtr := repository.GetWhereCondition(db.ColRecordsFinishTime, query.LowerThan, db.LessThan)
	if lowerThanWhereConPtr != nil {
		out = append(out, *lowerThanWhereConPtr)
	}

	return out
}

//Helper to get table for querying race records.
//Needs player to search on player name and races-racecatgamecat-racecategories to search on race category names
func getRecordQueryTable() string {
	table := db.JoinTables(db.TableRecords, db.TablePlayers, db.ColRecordsPlayerID, db.ColPlayerID)
	table = db.JoinTables(table, db.TableRaces, db.ColRecordsRaceID, db.ColRaceID)
	table = db.JoinTables(table, db.TableRaceCatGameCat, db.ColRaceCategoryID, db.ColRaceCatGameCatRaceCategoryID)
	table = db.JoinTables(table, db.TableRaceCategories, db.ColRaceCatGameCatRaceCategoryID, db.ColRaceCategoryID)

	return table
}

//Helper to parse DB query into records
func parseRecordQuery(database *sql.DB, res map[string][]any) []*Record {
	out := make([]*Record, 0) //slice of record IDs to parse runs

	//Make records for each result
	for i := range len(res[db.ColRecordID]) {
		//Parse required values
		playerName, ok := res[db.ColPlayerName][i].(string)
		if !ok {
			continue
		}

		raceID, ok := res[db.ColRecordsRaceID][i].(int64)
		if !ok {
			continue
		}

		numCollected, ok := res[db.ColRecordsNumCollected][i].(int64)
		if !ok {
			continue
		}

		recordID, ok := res[db.ColRecordID][i].(int64)
		if !ok {
			continue
		}

		//Unrequired values
		finishTime := res[db.ColRecordsFinishTime][i]

		//Get player from name
		player, err := players.GetPlayerByName(database, repository.MakeNullableStr(playerName))
		if err != nil {
			continue
		}

		//Get race from ID
		race, err := races.GetRaceByID(database, raceID)
		if err != nil {
			continue
		}

		//Get new record
		record := &Record{
			Player: player,
			Race: race,
			FinishTime: repository.MakeNullableStr(finishTime),
			NumCollected: repository.MakeNullableInt(numCollected),
			RecordID: recordID,
		}

		out = append(out, record)
	}

	return out
}

