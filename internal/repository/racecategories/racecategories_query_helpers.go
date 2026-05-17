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
	on := db.GetOnClause(db.TableRaceCategories, db.TableRaceCatGameCat, db.ColRaceCategoryID, db.ColRaceCatGameCatRaceCategoryID)
	table := db.JoinTables(db.TableRaceCategories, db.TableRaceCatGameCat, on)

	on = db.GetOnClause(db.TableRaceCatGameCat, db.TableGameCategories, db.ColRaceCatGameCatGameCatgeoryID, db.ColGameCategoryID)
	table = db.JoinTables(table, db.TableGameCategories, on)

	on = db.GetOnClause(db.TableGameCategories, db.TableGames, db.ColGameCategoryGameID, db.ColGameID)
	table = db.JoinTables(table, db.TableGames, on)

	return table
}

//Interprets results from query
func parseRaceCategoryQueryResults(database *sql.DB, res map[string][]any) ([]*RaceCategory, error) {
	out := make([]*RaceCategory, 0)

	//Cache for race categories we've gotten. Since they have a 1:M relationship, multiple race categories will be returned with different game categories
	//{raceID: *RaceCategory}
	catCache := make(map[string]*RaceCategory)
	for i := range len(res[db.ColRaceCategoryID]) {
		//Get required values
		raceCatID, ok := res[db.ColRaceCategoryID][i].(int64)
		if !ok {
			continue
		}

		raceCatName, ok := res[db.ColRaceCategoryName][i].(string)
		if !ok {
			continue
		}

		//Skip this game category if we've already seen it. Left join will return more race categories, so filter duplicates
		if _, exists := catCache[raceCatName]; !exists {
			newRaceCat := &RaceCategory{
				Name: repository.MakeNullableStr(raceCatName),
				CategoryID: raceCatID,
				GameCategories: make([]*gamecategories.GameCategory, 0),
			}
			catCache[raceCatName] = newRaceCat
		} //Race category hasn't been seen yet
	}

	//Get game categories for each race category
	for k, v := range catCache {
		gameCats, err := gamecategories.GetGameCategoriesFromRaceCategory(database, k)
		if err != nil {
			return nil, err
		}
		v.GameCategories = append(v.GameCategories, gameCats...)
	}

	//Add cached values to slice
	for _, v := range catCache {
		out = append(out, v)
	}

	return out, nil
}