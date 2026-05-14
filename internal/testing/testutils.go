package testutils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

//Package for general testing functionality

type HandlerTest struct {
	TestName string
	Body map[string]any
	RequestType string
	URL string
	ExpectedReponseCode int
	ExpectedSuccess bool
}

//Takes a test and a handler to call and returns the response decoded as a map
func CallHandler(t *testing.T, test HandlerTest, handlerFunc func(http.ResponseWriter, *http.Request)) map[string]any {
	t.Helper()

	//Encode request body
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(test.Body)
	if err != nil {
		t.Fatalf("%s: failed to encode request: %v", test.TestName, err)
	}

	req := httptest.NewRequest(test.RequestType, test.URL, buf) //Request

	//Call handler and decode response
	res := httptest.NewRecorder() //ResponseRecorder
	handlerFunc(res, req)

	var res_map map[string]any
	err = json.NewDecoder(res.Body).Decode(&res_map)
	if err != nil {
		t.Fatalf("%s: failed to decode json response: %v", test.TestName, err)
	}

	if res.Code != test.ExpectedReponseCode {
		t.Errorf("%s: incorrect reponse code. expected %v, got %v", test.TestName, test.ExpectedReponseCode, res.Code)
	}

	success, ok := res_map["success"].(bool)
	if !ok {
		t.Fatalf("%s: unable to parse success field as boolean", test.TestName)
	}

	if success != test.ExpectedSuccess {
		t.Errorf("%s: returned success doesn't match expected value. expected %v, got %v", test.TestName, test.ExpectedSuccess, success)
	}

	return res_map
}