package records

import (
	"database/sql"

	"github.com/multimario_api/internal/db"
	"github.com/multimario_api/internal/repository"
	"github.com/multimario_api/internal/repository/players"
	"github.com/multimario_api/internal/repository/races"
	"github.com/multimario_api/internal/twitch"
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
	raceIDWhereConPtr := repository.GetWhereCondition(db.TableRecords + "." + db.ColRecordsRaceID, query.RaceIDs, db.Equals)
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
	on := db.GetOnClause(db.TableRecords, db.TablePlayers, db.ColRecordsPlayerID, db.ColPlayerID)
	table := db.JoinTables(db.TableRecords, db.TablePlayers, on)

	on = db.GetOnClause(db.TablePlayers, db.TableSocials, db.ColPlayerID, db.ColSocialsPlayerID)
	table = db.JoinTables(table, db.TableSocials, on)

	on = db.GetOnClause(db.TableRecords, db.TableRaces, db.ColRecordsRaceID, db.ColRaceID)
	table = db.JoinTables(table, db.TableRaces, on)

	on = db.GetOnClause(db.TableRaces, db.TableRaceCatGameCat, db.ColRaceCategoryID, db.ColRaceCatGameCatRaceCategoryID)
	table = db.JoinTables(table, db.TableRaceCatGameCat, on)

	on = db.GetOnClause(db.TableRaceCatGameCat, db.TableRaceCategories, db.ColRaceCatGameCatRaceCategoryID, db.ColRaceCategoryID)
	table = db.JoinTables(table, db.TableRaceCategories, on)

	return table
}

//Helper to parse DB query into records
func parseRecordQuery(database *sql.DB, res map[string][]any) []*Record {
	out := make([]*Record, 0) //slice of record IDs to parse runs
	twitchIDCache := make(map[string]string) //Cache of {id: twich name}
	twitchIDs := make([]string, 0)

	//Get twitch IDs for each player
	for i := range len(res[db.ColRecordID]) {
		twitchID, ok := res[db.ColSocialsPlatformUserID][i].(string)
		if !ok {
			continue
		}
		twitchIDs = append(twitchIDs, twitchID)
	}

	if len(twitchIDs) != 0 {
		twitchIDCache, _ = twitch.GetTwitchNamesBatched(twitchIDs)
	}

	//Loop back through results to add stuff
	for i := range len(res[db.ColRecordID]) {
		//Parse required values
		playerName, ok := res[db.ColPlayerName][i].(string)
		if !ok {
			continue
		}

		playerID, ok := res[db.ColPlayerID][i].(int64)
		if !ok {
			continue
		}

		twitchID, ok := res[db.ColSocialsPlatformUserID][i].(string)
		if !ok {
			continue
		}

		twitchName := twitchIDCache[twitchID]

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

		//Get race from ID
		race, err := races.GetRaceByID(database, raceID)
		if err != nil {
			continue
		}

		//Get new record
		player := &players.Player{
			Name: repository.MakeNullableStr(playerName),
			TwitchName: repository.MakeNullableStr(twitchName),
			PlayerID: playerID,
		}
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

