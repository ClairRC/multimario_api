package twitch

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

//Package for handling Twitch integration.
//For the backend, this just gets Twitch IDs from Twitch names, and verifies their existence

//Interface to be able to mock twitch client for testing purposes
type TwitchAPICaller interface {
	GetTwitchIDFromName(string) (string, error)
	GetTwitchNameFromID(string) (string, error)
	GetTwitchIDFromToken(string) (string, error)
}

//Struct for twitch API calls
type TwitchClient struct {}

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

//Default twitch client for API calls since the actual struct doesn't have any state, its just there for the interface.
var Client TwitchAPICaller = TwitchClient{}

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
func (c TwitchClient) GetTwitchIDFromName(twitchName string) (string, error) {
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

//Gets twitch IDs from a list of names instead of just 1
//Returns slices of map in format {login name: id}
func GetTwitchIDsBatched(twitchNames []string) (map[string]string, error) {
	//Get number of requests necessary
	out := make(map[string]string)

	for i := 0; i < len(twitchNames); i+=100 {
		params := url.Values{}
		end := min(i+100, len(twitchNames))
    	batch := twitchNames[i:end]
		for _, name := range batch {
			params.Add("login", strings.ToLower(name))
		}
		endpoint := "https://api.twitch.tv/helix/users?" + params.Encode()

		//Create request
		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			return nil, err
		}
		authHeader := fmt.Sprintf("Bearer %s", appAccessToken)
		req.Header.Set("Authorization", authHeader)
		req.Header.Set("Client-Id", clientID)

		//Get response
		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			refreshAccessToken()
			authHeader = fmt.Sprintf("Bearer %s", appAccessToken)
			req.Header.Set("Authorization", authHeader)
			resp, err = httpClient.Do(req)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusUnauthorized {
				return nil, AccessTokenInvalidErr
			}
		}
		if resp.StatusCode == http.StatusBadRequest {
			return nil, UserCouldNotBeFoundErr
		}

		//Parse the response
		var twitchUserResp TwitchUserResponse
		err = json.NewDecoder(resp.Body).Decode(&twitchUserResp)
		if err != nil {
			return nil, errors.New("unknown error parsing twitch response. could not parse as json")
		}

		for _, user := range twitchUserResp.Data {
			out[user.Login] = user.ID
		}
	}

	return out, nil
}

//Get twitch name from twitch ID
func (c TwitchClient) GetTwitchNameFromID(twitchID string) (string, error){
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

//Gets twitch names from a list of IDs instead of just 1
//Returns slices of map in format {id: login name}
func GetTwitchNamesBatched(twitchIDs []string) (map[string]string, error) {
	//Get number of requests necessary
	out := make(map[string]string)

	for i := 0; i < len(twitchIDs); i+=100 {
		params := url.Values{}
		end := min(i+100, len(twitchIDs))
    	batch := twitchIDs[i:end]
		for _, id := range batch {
			params.Add("id", id)
		}
		endpoint := "https://api.twitch.tv/helix/users?" + params.Encode()

		//Create request
		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			return nil, err
		}
		authHeader := fmt.Sprintf("Bearer %s", appAccessToken)
		req.Header.Set("Authorization", authHeader)
		req.Header.Set("Client-Id", clientID)

		//Get response
		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			refreshAccessToken()
			authHeader = fmt.Sprintf("Bearer %s", appAccessToken)
			req.Header.Set("Authorization", authHeader)
			resp, err = httpClient.Do(req)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusUnauthorized {
				return nil, AccessTokenInvalidErr
			}
		}
		if resp.StatusCode == http.StatusBadRequest {
			return nil, UserCouldNotBeFoundErr
		}

		//Parse the response
		var twitchUserResp TwitchUserResponse
		err = json.NewDecoder(resp.Body).Decode(&twitchUserResp)
		if err != nil {
			return nil, errors.New("unknown error parsing twitch response. could not parse as json")
		}

		for _, user := range twitchUserResp.Data {
			out[user.ID] = user.Login
		}
	}

	return out, nil
}

//Get twitch ID from user token
func (c TwitchClient) GetTwitchIDFromToken(token string) (string, error){
	//Create request
	req, err := http.NewRequest("GET", "https://api.twitch.tv/helix/users", nil)
	if err != nil {
		return "", err
	}

	authHeader := fmt.Sprintf("Bearer %s", token)
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Client-Id", clientID)

	//Get response
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	//Check for errors
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("unknown error getting user info from twitch: " + http.StatusText(resp.StatusCode)) //TODO: Could be more specific based on response code
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

//Gets user token from Auth code from twitch
func GetUserTokenFromCode(code string, redirectURI string) (string, error) {
	//Send POST request to Twitch API to get user token
	params := url.Values{}
	params.Set("client_id", clientID)
	params.Set("client_secret", clientSecret)
	params.Set("code", code)
	params.Set("grant_type", "authorization_code")
	params.Set("redirect_uri", redirectURI)
	endpoint := "https://id.twitch.tv/oauth2/token?" + params.Encode()

	//Send POST request and parse body into JSON
	twitchResp, err := http.Post(endpoint, "application/x-www-form-urlencoded", bytes.NewReader(make([]byte, 0)))
	if err != nil {
		return "", errors.New("unable to send post request to twitch")
	}
	defer twitchResp.Body.Close()

	var resp TwitchTokenResponse
	err = json.NewDecoder(twitchResp.Body).Decode(&resp)
	if err != nil {
		return "", errors.New("unable to decode twitch response as json")
	}

	//Return the token
	return resp.AccessToken, nil
}

//Redirects a user to Twitch to retrieve an access token
func GetUserTokenRedirectURL(callbackURI string) *url.URL {
	redirURL, _ := url.Parse("https://id.twitch.tv/oauth2/authorize")

	params := redirURL.Query()
	params.Set("client_id", clientID)
	params.Set("redirect_uri", callbackURI)
	params.Set("response_type", "code")
	params.Set("scope", "user:read:email") //Gives access to /users endpoint

	redirURL.RawQuery = params.Encode()

	return redirURL
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

//Sets twitch Client
//Useful for mocking twitch API calls for testing
func SetTwitchClient(client TwitchAPICaller) {
	//TODO: This might not be a super clean solution but like. It works I just don't want real API calls in testing.
	Client = client
}