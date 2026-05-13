package runs

import (
	"github.com/multimario_api/internal/db"
	"github.com/multimario_api/internal/repository"
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

	//TODO: Finish
}
