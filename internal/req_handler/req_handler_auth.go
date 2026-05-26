package req_handler

import (
	"net/http"
	"strings"

	"github.com/multimario_api/internal/auth"
	"github.com/multimario_api/internal/twitch"
)

/*
* File for authentication middleware and request handlers
 */

//Middleware that authenticates user api  key
func (h *ReqHandler) Authenticate(next http.HandlerFunc, level auth.AuthLevel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//Get API key from header
		key := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")

		//If key is empty and auth level is over none, this is unauthorized
		if key == "" && level > auth.AuthNone {
			writeError(w, http.StatusUnauthorized, "unauthorized: must include an api key for this resource")
			return
		}

		valid, err := auth.KeyMeetsLevel(h.DataBase, key, level)
		if err != nil || !valid {
			writeError(w, http.StatusBadRequest, "key not present in database")
			return
		}

		//Key not valid
		if !valid {
			writeError(w, http.StatusForbidden, "key does not meet authentication level to access this endpoint")
			return
		}

		//Key is valid, log this access and call the next function the next function
		next(w, r)
	}
}

//Handler that a user calls to create an API key
func (h *ReqHandler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	//Get Twitch redirect parameters
	rURL := twitch.GetUserTokenRedirectURL("http://localhost:8080/auth/twitch/callback") //TODO: Put the URI somewhere else thats easier to change

	http.Redirect(w, r, rURL.String(), http.StatusTemporaryRedirect)
}

//Twitch callback function for getting Twitch auth code
func (h *ReqHandler) TwitchCallback(w http.ResponseWriter, r *http.Request) {
	//Get callback parameters
	urlCode := r.URL.Query()["code"]
	urlError := r.URL.Query()["error"]
	
	//If there's an error, unable to get an API key
	if len(urlError) > 0 {
		writeError(w, http.StatusUnauthorized, "user must authorize with twitch to generate an api key")
		return
	}

	//Get code and swap it for token
	code := urlCode[0]

	token, err := twitch.GetUserTokenFromCode(code, "http://localhost:8080/auth/twitch/callback")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error: " + err.Error())
		return
	}

	twitchID, err := twitch.Client.GetTwitchIDFromToken(token)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unknown error getting twitch id: " + err.Error())
		return
	}

	//Get API key
	apiKey, err := auth.GetAPIKey(h.DataBase, twitchID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unknown error getting api key: " + err.Error())
		return
	}

	//Write key
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "api_key": apiKey}, nil)
}