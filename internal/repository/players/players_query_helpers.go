package players

import (
	"github.com/multimario_api/internal/db"
	"github.com/multimario_api/internal/repository"
	"github.com/multimario_api/internal/twitch"
)

//This file includes helper functions for parsing Player DB queries

func getPlayerWhereCons(playerQuery PlayerQuery, twitchIDCache map[string]string) ([]db.WhereCondition, error) {
	out := make([]db.WhereCondition, 0) //Return value

	//Get name where conditions
	var nameWherePtr *db.WhereCondition
	for i, name := range playerQuery.Names {
		if i == 0 {
			nameWherePtr = &db.WhereCondition{
				ColName: db.ColPlayerName,
				Op: db.Equals,
				Value: name,
				Ors: make([]db.OrCondition, 0),
			}
		} else {
			nameWherePtr.Ors = append(nameWherePtr.Ors, db.OrCondition{
				Op: db.Equals,
				Value: name,
			})
		}
	}
	if nameWherePtr != nil {
		out = append(out, *nameWherePtr)
	}

	//Get twitch name where conditions
	//Get Twitch IDs from twitch names
	var twitchIDWherePtr *db.WhereCondition
	for i, twitchName := range playerQuery.TwitchNames {
		//Get twitch ID from the name
		id, err := twitch.GetTwitchIDFromName(twitchName)
		if err != nil {
			return nil, err
		}
		twitchIDCache[id] = twitchName //Add ID to cache

		if i == 0 {
			twitchIDWherePtr = &db.WhereCondition{
				ColName: db.ColSocialsPlatformUserID,
				Op: db.Equals,
				Value: id,
				Ors: make([]db.OrCondition, 0),
			}
		} else {
			twitchIDWherePtr.Ors = append(twitchIDWherePtr.Ors, db.OrCondition{
				Op: db.Equals,
				Value: id,
			})
		}
	}
	if twitchIDWherePtr != nil {
		out = append(out, *twitchIDWherePtr)
	}

	return out, nil
}

//Parses DB response into player slice
func parsePlayerQueryResponse(res map[string][]any, twitchIDCache map[string]string) ([]*Player, error) {
	out := make([]*Player, 0)

	//Loop through results and create players
	for i := range len(res[db.ColPlayerID]) {
		//Make sure values are valid else don't include in the response
		name, ok := res[db.ColPlayerName][i].(string)
		if !ok {
			continue
		}

		twitchID, ok := res[db.ColSocialsPlatformUserID][i].(string)
		if !ok {
			continue
		}

		twitchNameStr, cached := twitchIDCache[twitchID] //Check if twitch ID is cached
		if !cached {
			alsoTwitchNameStr, err := twitch.GetTwitchNameFromID(twitchID)
			twitchNameStr = alsoTwitchNameStr //Good name
			if err != nil {
				return nil, err
			} //If not cache, call API for twitch name
		}

		twitchName := repository.MakeNullableStr(twitchNameStr) //Can be nullable since this is not technically required required

		id, ok := res[db.ColPlayerID][i].(int64)
		if !ok {
			continue
		}

		newPlayer := &Player {
			Name: repository.MakeNullableStr(name),
			TwitchName: twitchName,
			PlayerID: id,
		}
		out = append(out, newPlayer)
	}

	return out, nil
}