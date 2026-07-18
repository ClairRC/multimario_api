package players

import (
	"errors"
	"strings"

	"github.com/multimario_api/internal/db"
	"github.com/multimario_api/internal/repository"
	"github.com/multimario_api/internal/twitch"
)

//This file includes helper functions for parsing Player DB queries

func getPlayerWhereCons(playerQuery PlayerQuery, twitchIDCache map[string]string) ([]db.WhereCondition, error) {
	out := make([]db.WhereCondition, 0) //Return value

	//Get where conditions
	nameWherePtr := repository.GetWhereCondition(db.ColPlayerName, playerQuery.Names, db.Equals)
	if nameWherePtr != nil {
		out = append(out, *nameWherePtr)
	}

	//Get Twitch IDs for each player
	idMap, err := twitch.GetTwitchIDsBatched(playerQuery.TwitchNames)
	if err != nil {
		return nil, err
	}

	//Get twitch name where conditions
	//This caches twitch IDs so the logic is too specific for the general GetWhereCondition helper
	var twitchIDWherePtr *db.WhereCondition
	for i, twitchName := range playerQuery.TwitchNames {
		//Get twitch ID from the name
		id, exists := idMap[strings.ToLower(twitchName)]
		if !exists {
			return nil, errors.New("unknown error trying to get twitch ids.")
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

	//Loop through results and fill cache to get missing players
	newIDQueue := make([]string, 0)
	for i := range len(res[db.ColPlayerID]) {
		twitchID, ok := res[db.ColSocialsPlatformUserID][i].(string)
		if !ok {
			continue
		}

		_, cached := twitchIDCache[twitchID] //Check if twitch ID is cached
		//If not in cache, add it to the newID queue
		if !cached {
			newIDQueue = append(newIDQueue, twitchID)
		}
	}

	//Fill cache with new values
	if len(newIDQueue) > 0 {
		newCachedValues, err := twitch.GetTwitchNamesBatched(newIDQueue)
		if err != nil {
			return nil, err
		}

		for k, v := range newCachedValues{
			twitchIDCache[k] = v
		}
	}

	//Loop back though results to add to the output
	for i := range len(res[db.ColPlayerID]) {
		//Make sure values are valid else don't include in the response
		name, ok := res[db.ColPlayerName][i].(string)
		if !ok {
			continue
		}

		id, ok := res[db.ColPlayerID][i].(int64)
		if !ok {
			continue
		}

		twitchID, ok := res[db.ColSocialsPlatformUserID][i].(string)
		if !ok {
			continue
		}

		newPlayer := &Player {
			Name: repository.MakeNullableStr(name),
			TwitchName: repository.MakeNullableStr(twitchIDCache[twitchID]),
			PlayerID: id,
		}
		out = append(out, newPlayer)
	}

	return out, nil
}