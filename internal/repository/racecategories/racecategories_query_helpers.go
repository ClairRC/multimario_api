package racecategories

import (
	"database/sql"

	"github.com/multimario_api/internal/db"
	"github.com/multimario_api/internal/repository"
	"github.com/multimario_api/internal/repository/gamecategories"
)

//Helpers for querying race categories

//Gets race category where conditions
func getRaceCategoryWhereCons(query RaceCategoryQuery) []db.WhereCondition {
	out := make([]db.WhereCondition, 0)

	//Race cat names
	raceCatNameWherePtr := repository.GetWhereCondition(db.ColRaceCategoryName, query.RaceCategories, db.Equals)
	if raceCatNameWherePtr != nil {
		out = append(out, *raceCatNameWherePtr)
	}

	//Game names
	gameNameWherePtr := repository.GetWhereCondition(db.ColGameName, query.Games, db.Equals)
	if gameNameWherePtr != nil {
		out = append(out, *gameNameWherePtr)
	}

	//Game category names
	gameCatWherePtr := repository.GetWhereCondition(db.ColGameCategoryName, query.GameCategories, db.Equals)
	if gameCatWherePtr != nil {
		out = append(out, *gameCatWherePtr)
	}

	return out
}

//Needs racecategories-racecatgamecat-gamecatories for querying by game categories and games for querying by game names
func getRaceCategoryQueryTable() string {
	table := db.JoinTables(db.TableRaceCategories, db.TableRaceCatGameCat, db.ColRaceCategoryID, db.ColRaceCatGameCatRaceCategoryID)
	table = db.JoinTables(table, db.TableGameCategories, db.ColRaceCatGameCatGameCatgeoryID, db.ColGameCategoryID)
	table = db.JoinTables(table, db.TableGames, db.ColGameCategoryGameID, db.ColGameID)

	return table
}

//Interprets results from query
func parseRaceCategoryQueryResults(database *sql.DB, res map[string][]any) ([]*RaceCategory, error) {
	out := make([]*RaceCategory, 0)

	//Cache for race categories we've gotten. Since they have a 1:M relationship, multiple race categories will be returned with different game categories
	//{raceID: *RaceCategory}
	catCache := make(map[int64]*RaceCategory)
	for i := range len(res[db.ColRaceCategoryID]) {
		//Get required values
		raceCatID, ok := res[db.ColRaceCategoryID][i].(int64)
		if !ok {
			continue
		}

		gameCategoryID, ok := res[db.ColGameCategoryID][i].(int64)
		if !ok {
			continue
		}

		raceCatName, ok := res[db.ColRaceCategoryName][i].(string)
		if !ok {
			continue
		}

		//Get the category from the ID
		//This is an extra couple of DB calls, but there's only a handful of race categories and game categories per race category
		//So it's not worth the extra complexity to do it all in 1 call currently.
		gameCat, err := gamecategories.GetGameCategoryByID(database, gameCategoryID)
		if err != nil {
			continue
		}

		//Check if we've already seen this race category. If so, create a new one
		if _, exists := catCache[raceCatID]; !exists {
			newRaceCat := &RaceCategory{
				Name: repository.MakeNullableStr(raceCatName),
				CategoryID: raceCatID,
				GameCategories: make([]*gamecategories.GameCategory, 0),
			}
			catCache[raceCatID] = newRaceCat
		} //Race category hasn't been seen yet

		//Add game category to the race category list
		catCache[raceCatID].GameCategories = append(catCache[raceCatID].GameCategories, gameCat)
	}

	//Add cached values to slice
	for _, v := range catCache {
		out = append(out, v)
	}

	return out, nil
}