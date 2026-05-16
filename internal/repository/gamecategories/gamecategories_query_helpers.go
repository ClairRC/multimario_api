package gamecategories

import (
	"github.com/multimario_api/internal/db"
	"github.com/multimario_api/internal/repository"
	"github.com/multimario_api/internal/repository/games"
)

//File for holding helper functions for parsing query results

//Gets the table that is used for the GameCategoryQuery.
//We need GameCategories, Games, GameCatRaceCat, and RaceCategories for the query
func getGameCategoryQueryTable() string {
	//Note: GameCatRaceCat is only necessary for searching on RaceCategory name. We don't actually need information from that table.
	on := db.GetOnClause(db.TableGameCategories, db.TableGames, db.ColGameCategoryGameID, db.ColGameID)
	table := db.JoinTables(db.TableGameCategories, db.TableGames, on)

	on = db.GetOnClause(db.TableGameCategories, db.TableRaceCatGameCat, db.ColGameCategoryID, db.ColRaceCatGameCatGameCatgeoryID)
	table = db.JoinTables(table, db.TableRaceCatGameCat, on)

	on = db.GetOnClause(db.TableRaceCatGameCat, db.TableRaceCategories, db.ColRaceCatGameCatRaceCategoryID, db.ColRaceCategoryID)
	table = db.JoinTables(table, db.TableRaceCategories, on)

	return table
}

//Gets where conditions for game category
func getGameCategoryWhereCons(gamecategoryQuery GameCategoryQuery) []db.WhereCondition {
	out := make([]db.WhereCondition, 0)

	//Query for names
	catNameWherePtr := repository.GetWhereCondition(db.ColGameCategoryName, gamecategoryQuery.Names, db.Equals)
	if catNameWherePtr != nil {
		out = append(out, *catNameWherePtr)
	}

	//Query for race categories
	raceCatNameWherePtr := repository.GetWhereCondition(db.ColRaceCategoryName, gamecategoryQuery.RaceCategories, db.Equals)
	if raceCatNameWherePtr != nil {
		out = append(out, *raceCatNameWherePtr)
	}

	//Query for game names
	gameNameWherePtr := repository.GetWhereCondition(db.ColGameName, gamecategoryQuery.GameNames, db.Equals)
	if gameNameWherePtr != nil {
		out = append(out, *gameNameWherePtr)
	}

	return out
}

//Function to parse query results
func parseGameCategoryQueryResults(res map[string][]any) []*GameCategory {
	out := make([]*GameCategory, 0)

	for i := range len(res[db.ColGameCategoryID]) {
		//Make sure types of required values are valid
		gameID, ok := res[db.ColGameCategoryGameID][i].(int64)
		if !ok {
			continue
		}
		gameName, ok := res[db.ColGameName][i].(string)
		if !ok {
			continue
		}
		catID, ok := res[db.ColGameCategoryID][i].(int64)
		if !ok {
			continue
		}
		catName, ok := res[db.ColGameCategoryName][i].(string)
		if !ok {
			continue
		}
		numCollectibles, ok := res[db.ColGameCategoryNumCollectibles][i].(int64)
		if !ok {
			continue
		}

		estimate := res[db.ColGameCategoryEstimate][i] //No type assertion because this is allowed to be NULL

		game := &games.Game{
			GameID: gameID,
			Name: repository.MakeNullableStr(gameName),
		}
		category := &GameCategory{
			Name: repository.MakeNullableStr(catName),
			Estimate: repository.MakeNullableStr(estimate),
			NumCollectibles: repository.MakeNullableInt(numCollectibles),
			CategoryID: catID,
			Game: game,
		}

		out = append(out, category)
	}

	return out
}