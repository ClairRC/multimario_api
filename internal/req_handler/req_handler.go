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

	//Add logic that this can't be less than 1
	if pageNum < 1 {
		pageNum = 1
	}

	return pageNum, nil
}

//Function to convert items into paginated items. Returns metadata for this query
func getPaginationMetadata(totalResultCount int64, url *url.URL, pageNum int, limit int) (map[string]any) {
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

	if offset + limit < int(totalResultCount) {
		urlCopy := *url
		q := urlCopy.Query()
		q.Set("page_num", strconv.Itoa(pageNum + 1))
		urlCopy.RawQuery = q.Encode()
		meta["next_url"] = urlCopy.String()
	} else {
		meta["next_url"] = nil
	}
	meta["total_items"] = totalResultCount

	return meta
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

//Adds up all the times passed in for calculating run estimates
func addTimes(times []repository.NullableStr) (repository.NullableStr, error) {
	var totalSeconds int64

	//If times is empty, just return 00:00:00 immediately
	if len(times) == 0 {
		return repository.MakeNullableStr("00:00:00"), nil
	}

	//Flag to check if the sum is NULL for any reason
	nullSum := true
	for _, tStr := range times {
		if !tStr.Valid {
			continue //Skip NULL strings
		}

		//At least 1 time is valid, so the sum is not null
		nullSum = false

		t := tStr.Value
		parts := strings.Split(t, ":")
		if len(parts) != 3 {
			return repository.NULLStr, fmt.Errorf("invalid time format: %q (expected hh:mm:ss)", t)
		}

		var h, m, s int
		_, err := fmt.Sscanf(t, "%d:%d:%d", &h, &m, &s)
		if err != nil {
			return repository.NULLStr, fmt.Errorf("invalid time format: %q: %w", t, err)
		}

		if m < 0 || m > 59 || s < 0 || s > 59 {
			return repository.NULLStr, fmt.Errorf("invalid time value: %q (minutes/seconds out of range)", t)
		}

		totalSeconds += int64(h)*3600 + int64(m)*60 + int64(s)
	}

	//If none of the times were non-null, return null string
	if nullSum {
		return repository.NULLStr, nil
	}

	//Otherwise, calculate actual result
	
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	return repository.MakeNullableStr(fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)), nil
}