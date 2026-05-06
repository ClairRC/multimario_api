package twitch

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

//Package for handling Twitch integration.
//For the backend, this just gets Twitch IDs from Twitch names, and verifies their existence

type TwitchTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn int `json:"expires_in"`
	TokenType string `json:"token_type"`
}

type TwitchUserResponse struct {
	Data []TwitchUser `json:"data"`
}

type TwitchUser struct {
	ID string `json:"id"`
	Login string `json:"login"`
}

var httpClient *http.Client = &http.Client{Timeout: 10 * time.Second}
var clientID string = ""
var clientSecret string = ""
var appAccessToken string = "" //App access token

//Default errors
var UserCouldNotBeFoundErr error = errors.New("unable to get user")
var AccessTokenInvalidErr error = errors.New("could not get valid access token")

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
		return errors.New("unable to decode twitch response as json")
	}

	//Set values
	clientID = twitchClientID
	clientSecret = twitchClientSecret
	appAccessToken = resp.AccessToken
	
	return nil
}

//Gets Twitch ID from name. Returns an error if unable to get it
func GetTwitchIDFromName(twitchName string) (string, error) {
	params := url.Values{}
	params.Set("login", twitchName)
	endpoint := "https://api.twitch.tv/helix/users?" + params.Encode()

	//Create request
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}
	authHeader := fmt.Sprintf("Bearer %s", appAccessToken)
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Client-Id", clientID)

	//Get response
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		refreshAccessToken()
		authHeader = fmt.Sprintf("Bearer %s", appAccessToken)
		req.Header.Set("Authorization", authHeader)
		resp, err = httpClient.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			return "", AccessTokenInvalidErr
		}
	}
	if resp.StatusCode == http.StatusBadRequest {
		return "", UserCouldNotBeFoundErr
	}

	//Parse the response
	var twitchUserResp TwitchUserResponse
	err = json.NewDecoder(resp.Body).Decode(&twitchUserResp)
	if err != nil {
		return "", errors.New("unknown error parsing twitch response. could not parse as json")
	}

	if len(twitchUserResp.Data) == 0 {
		return "", UserCouldNotBeFoundErr
	}

	return twitchUserResp.Data[0].ID, nil
}

//Get twitch name from twitch ID
func GetTwitchNameFromID(twitchID string) (string, error){
	params := url.Values{}
	params.Set("id", twitchID)
	endpoint := "https://api.twitch.tv/helix/users?" + params.Encode()

	//Create request
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}
	authHeader := fmt.Sprintf("Bearer %s", appAccessToken)
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Client-Id", clientID)

	//Get response
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		refreshAccessToken()
		authHeader = fmt.Sprintf("Bearer %s", appAccessToken)
		req.Header.Set("Authorization", authHeader)
		resp, err = httpClient.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			return "", AccessTokenInvalidErr
		}
	}
	if resp.StatusCode == http.StatusBadRequest {
		return "", UserCouldNotBeFoundErr
	}

	//Parse the response
	var twitchUserResp TwitchUserResponse
	err = json.NewDecoder(resp.Body).Decode(&twitchUserResp)
	if err != nil {
		return "", errors.New("unknown error parsing twitch response. could not parse as json")
	}

	if len(twitchUserResp.Data) == 0 {
		return "", UserCouldNotBeFoundErr
	}

	return twitchUserResp.Data[0].Login, nil
}

//Gets new access token from twitch
func refreshAccessToken() {
	//Send POST request to Twitch API to get access token
	params := url.Values{}
	params.Set("client_id", clientID)
	params.Set("client_secret", clientSecret)
	params.Set("grant_type", "client_credentials")
	endpoint := "https://id.twitch.tv/oauth2/token?" + params.Encode()

	//Send POST request and parse body into JSON
	twitchResp, err := http.Post(endpoint, "application/x-www-form-urlencoded", bytes.NewReader(make([]byte, 0)))
	if err != nil {
		return
	}
	defer twitchResp.Body.Close()

	responseCode := twitchResp.StatusCode
	if responseCode > 299 {
		return
	}

	//Success, set new token
	var resp TwitchTokenResponse
	err = json.NewDecoder(twitchResp.Body).Decode(&resp)
	if err != nil {
		return
	}

	appAccessToken = resp.AccessToken
}