package req_handler

/*
* Package for handling requests to the backend. This is where the routing and transforming
* API calls into database calls should be.
 */

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

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
func writeJSON(w http.ResponseWriter, status int, data map[string]any, meta map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if meta != nil {
		data["meta"] = meta //Add metadata
	}
	json.NewEncoder(w).Encode(&data)
}

//Helper function to write error resposne
func writeError(w http.ResponseWriter, status int, err string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(&(map[string]any {"success": false, "error": err}))
}

//Function to parse URL parameters into page number for pagination
func getResponsePageNum(urlPageNumParams []string) (int, error) {
	//If there's no page number, defaults to 1
	pageNum := 1
	if len(urlPageNumParams) > 0 {
		parsed, err := strconv.Atoi(urlPageNumParams[0])
		if err != nil {
			return -1, err
		}

		if parsed > 0 {
			pageNum = parsed
		}
	}
	return pageNum, nil
}

//Function to convert items into paginated items. Returns paginated items and metadata
func paginate(items []map[string]any, url *url.URL, pageNum int, limit int) ([]map[string]any, map[string]any) {
	//TODO: This function is problematic. My get handlers get al ist of EVER row and THEN it just gets tossed out here
	//For now we're only looking at like maybe 1-2 seconds of waiting, but it's still very much not ideal.
	offset := (pageNum - 1) * limit

	//Get metadata for this
	meta := make(map[string]any)
	if pageNum > 1 {
		urlCopy := *url
		q := urlCopy.Query()
		q.Set("page_num", strconv.Itoa(pageNum-1))
		urlCopy.RawQuery = q.Encode()
		meta["prev_url"] = urlCopy.String()
	} else {
		meta["prev_url"] = nil
	}

	if offset + limit < len(items) {
		urlCopy := *url
		q := urlCopy.Query()
		q.Set("page_num", strconv.Itoa(pageNum + 1))
		urlCopy.RawQuery = q.Encode()
		meta["next_url"] = urlCopy.String()
	} else {
		meta["next_url"] = nil
	}
	meta["total_items"] = len(items)

	//Set items to only be between the two limits
	if offset >= len(items) {
		items = make([]map[string]any, 0)
	} else {
		items = items[offset:min(offset+limit, len(items))]
	}

	return items, meta
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

//Validate duration field
func validateDuration(w http.ResponseWriter, req map[string]any, key string, required bool) (repository.NullableStr, error) {
	err := (&DurationField{req[key]}).Validate()

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
			writeError(w, http.StatusBadRequest, key + " must be valid duration in hh:mm:ss format")
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
			writeError(w, http.StatusBadRequest, key + " must be valid time in hh:mm:ss format")
			return repository.NULLStr, err
		} 
	}

	return repository.MakeNullableStr(req[key].(string)), nil
}

//Validate date field
func validateDate(w http.ResponseWriter, req map[string]any, key string, required bool) (repository.NullableStr, error) {
	err := (&DateField{req[key]}).Validate()

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
			writeError(w, http.StatusBadRequest, key + " must be in YYYY-MM-DD format")
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

	//JSON decodes this as float64, so assert type as float64 instead of int. Weird but it works.
	return repository.MakeNullableInt(req[key].(float64)), nil
}

//Validates array field
func validateArray(w http.ResponseWriter, req map[string]any, key string, required bool) ([]any, error) {
	err := (&ArrayField{req[key]}).Validate()

	//Check error
	if err != nil {
		switch err{
		case FieldIsEmptyErr:
			if required {
				writeError(w, http.StatusBadRequest, key + " cannot be empty")
				return nil, err
			} else {
				return make([]any, 0), nil
			}
		case FieldIsWrongFormatErr:
			writeError(w, http.StatusBadRequest, key + " must be an array")
			return nil, err
		}
	}

	return req[key].([]any), nil
}

//Gets the sum of two strings that represent time in the format hh:mm:ss
func AddTimeStrings(a, b string) (string, error) {
    secondsA, err := timeStringToSeconds(a)
    if err != nil {
        return "", err
    }

    secondsB, err := timeStringToSeconds(b)
    if err != nil {
        return "", err
    }

    return secondsToTimeString(secondsA + secondsB), nil
}

func timeStringToSeconds(s string) (int, error) {
    parts := strings.Split(s, ":")
    if len(parts) != 3 {
        return 0, errors.New("expected format hh:mm:ss")
    }

    hours, err := strconv.Atoi(parts[0])
    if err != nil {
        return 0, err
    }
    minutes, err := strconv.Atoi(parts[1])
    if err != nil {
        return 0, err
    }
    seconds, err := strconv.Atoi(parts[2])
    if err != nil {
        return 0, err
    }

    if minutes < 0 || minutes > 59 || seconds < 0 || seconds > 59 {
        return 0, err
    }

    return hours*3600 + minutes*60 + seconds, nil
}

func secondsToTimeString(totalSeconds int) string {
    hours := totalSeconds / 3600
    minutes := (totalSeconds % 3600) / 60
    seconds := totalSeconds % 60

    return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}