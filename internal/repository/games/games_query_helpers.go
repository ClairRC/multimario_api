package games

import (
	"github.com/multimario_api/internal/db"
	"github.com/multimario_api/internal/repository"
)

//File of helper functions for parsing queries from GET requests

//Gets where conditions based on query instance
func getGameWhereCons(gameQuery GameQuery) []db.WhereCondition {
	out := make([]db.WhereCondition, 0) //Output slice

	//Get where condition
	newWhereConPtr := repository.GetWhereCondition(db.ColGameName, gameQuery.Names, db.Equals)
	if newWhereConPtr != nil {
		out = append(out, *newWhereConPtr)
	}

	return out
}

func parseQueryResponse(res map[string][]any) []*Game {
	out := make([]*Game, 0)
	
	for i := range len(res[db.ColGameID]) {
		//Don't include invalid responses
		name, ok := res[db.ColGameName][i].(string)
		if !ok {
			continue
		}
		id, ok := res[db.ColGameID][i].(int64)
		if !ok {
			continue
		}

		newGame := &Game{
			Name: repository.MakeNullableStr(name),
			GameID: id,
		}
		out = append(out, newGame)
	}
	return out
}