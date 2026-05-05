package twitch

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
)

//Package for handling Twitch integration.
//For the backend, this just gets Twitch IDs from Twitch names, and verifies their existence

type TwitchTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn int `json:"expires_in"`
	TokenType string `json:"token_type"`
}

var clientID string = ""
var clientSecret string = ""
var appAccessToken string = "" //App access token

//Sets the Twitch tokens and returns an error if unable to get access token
func SetTwitchParams(twitchClientID string, twitchClientSecret string) error {
	//Send POST request to Twitch API to get access token
	params := url.Values{}
	params.Set("client_id", twitchClientID)
	params.Set("client_secret", twitchClientSecret)
	params.Set("grant_type", "client_credentials")
	endpoint := "https://id.twitch.tv/oauth2/token?" + params.Encode()

	//Send POST request and parse body into JSON
	twitchResp, err := http.Post(endpoint, "application/x-www-form-urlencoded", bytes.NewReader(make([]byte, 0)))
	if err != nil {
		return errors.New("unable to send post request to twitch")
	}
	defer twitchResp.Body.Close()

	responseCode := twitchResp.StatusCode
	if responseCode > 299 {
		return errors.New("client_id or client_secret is invalid")
	}

	var resp TwitchTokenResponse
	err = json.NewDecoder(twitchResp.Body).Decode(&resp)
	if err != nil {
		return errors.New("unable to decode twitch reponse as json")
	}

	//Set values
	clientID = twitchClientID
	clientSecret = twitchClientSecret
	appAccessToken = resp.AccessToken
	
	return nil
}