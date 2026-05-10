package races

import (
	"database/sql"

	"github.com/multimario_api/internal/db"
	"github.com/multimario_api/internal/repository"
	"github.com/multimario_api/internal/repository/racecategories"
)

//File for helper functions that simplify query logic

//Gets race where conditions
func getRaceWhereCons(query RaceQuery) []db.WhereCondition {
	out := make([]db.WhereCondition, 0)

	//Get ID WhereConditions
	idWhereCon := repository.GetWhereCondition(db.ColRaceID, query.IDs, db.Equals)
	if idWhereCon != nil {
		out = append(out, *idWhereCon)
	}

	//Get BeforeDates where condition
	beforeWhereCon := repository.GetWhereCondition(db.ColRaceDate, query.BeforeDates, db.LessThan)
	if beforeWhereCon != nil {
		out = append(out, *beforeWhereCon)
	}

	//Get AfterDates where condition
	afterWhereCon := repository.GetWhereCondition(db.ColRaceDate, query.AfterDates, db.GreaterThan)
	if afterWhereCon != nil {
		out = append(out, *afterWhereCon)
	}

	//Get OnDates where condition
	onWhereCon := repository.GetWhereCondition(db.ColRaceDate, query.OnDates, db.Equals)
	if onWhereCon != nil {
		out = append(out, *onWhereCon)
	}

	//Get cateogories where condition
	catWhereCon := repository.GetWhereCondition(db.ColRaceCategoryName, query.Categories, db.Equals)
	if catWhereCon != nil {
		out = append(out, *catWhereCon)
	}

	//Get Status where condition
	statusWhereCon := repository.GetWhereCondition(db.ColRaceStatus, query.Statuses, db.Equals)
	if statusWhereCon != nil {
		out = append(out, *statusWhereCon)
	}

	return out
}

//Parses DB response
func parseRaceQueryResponse(database *sql.DB, res map[string][]any) ([]*Race, error){
	out := make([]*Race, 0)

	for i := range len(res[db.ColRaceID]) {
		//Verify required values
		id, ok := res[db.ColRaceID][i].(int64)
		if !ok {
			continue
		}
		status, ok := res[db.ColRaceStatus][i].(string)
		if !ok {
			continue
		}
		catName, ok := res[db.ColRaceCategoryName][i].(string)
		if !ok {
			continue
		}

		//Optional values
		date := res[db.ColRaceDate][i]
		start := res[db.ColRaceStartTime][i]

		//Get race category from db. This is more complicated to parse in 1 DB call than i care to optimize right now.
		//all the information can be gotten from 1 db call but it's too complicated for the size
		raceCat, err := racecategories.GetRaceCategoryByName(database, repository.MakeNullableStr(catName))
		if err != nil {
			return nil, err
		}

		//Race
		race := &Race{
			Date: repository.MakeNullableStr(date),
			StartTime: repository.MakeNullableStr(start),
			Status: repository.MakeNullableStr(status),
			RaceCategory: raceCat,
			RaceID: id,
		}
		out = append(out, race)
	}
	
	return out, nil
}