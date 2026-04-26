package req_handler

/*
* Package for handling requests to the backend. This is where the routing and transforming
* API calls into database calls should be.
 */

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/multimario_api/internal/repository"
)

//ReqHander struct to hold extra dependencies
type ReqHandler struct {
	DataBase *sql.DB
}

/*
* Helper functions for all Handler functions
*/

//Helper function to parse request body from JSON to map
func parseReqJSON(r *http.Request) (map[string]any, error) {
	var out map[string]any
	err := json.NewDecoder(r.Body).Decode(&out)
	if err != nil{
		return nil, err
	}

	return out, nil
}

//Helper function to write JSON to w
func writeJSON(w http.ResponseWriter, status int, data map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(&data)
}

//Helper function to write error resposne
func writeError(w http.ResponseWriter, status int, err string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(&(map[string]any {"success": false, "error": err}))
}

//Functions for validating field types

//Validate text field
func validateText(w http.ResponseWriter, req map[string]any, key string, required bool) (repository.NullableStr, error) {
	err := (&TextField{req[key]}).Validate()

	//Check error
	if err != nil {
		switch err{
		case FieldIsEmptyErr:
			if required {
				writeError(w, http.StatusBadRequest, key + " cannot be empty")
				return repository.NULLStr, err
			} else {
				return repository.NULLStr, nil
			}
		case FieldIsWrongFormatErr:
			writeError(w, http.StatusBadRequest, key + " must be string")
			return repository.NULLStr, err
		} 
	}

	return repository.MakeNullableStr(req[key].(string)), nil
}

//Validate time field
func validateTime(w http.ResponseWriter, req map[string]any, key string, required bool) (repository.NullableStr, error) {
	err := (&TimeField{req[key]}).Validate()

	//Check error
	if err != nil {
		switch err{
		case FieldIsEmptyErr:
			if required {
				writeError(w, http.StatusBadRequest, key + " cannot be empty")
				return repository.NULLStr, err
			} else {
				return repository.NULLStr, nil
			}
		case FieldIsWrongFormatErr:
			writeError(w, http.StatusBadRequest, key + " must be in hh:mm:ss format")
			return repository.NULLStr, err
		} 
	}

	return repository.MakeNullableStr(req[key].(string)), nil
}

//Validate number field
func validateNumber(w http.ResponseWriter, req map[string]any, key string, required bool) (repository.NullableInt, error) {
	err := (&IntField{req[key]}).Validate()

	//Check error
	if err != nil {
		switch err{
		case FieldIsEmptyErr:
			if required {
				writeError(w, http.StatusBadRequest, key + " cannot be empty")
				return repository.NULLInt, err
			} else {
				return repository.NULLInt, nil
			}
		case FieldIsWrongFormatErr:
			writeError(w, http.StatusBadRequest, key + " must be an int")
			return repository.NULLInt, err
		} 
	}

	return repository.MakeNullableInt(req[key].(int)), nil
}