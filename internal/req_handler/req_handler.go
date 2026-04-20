package req_handler

/*
* Package for handling requests to the backend. This is where the routing and transforming
* API calls into database calls should be.
 */

import (
	"database/sql"
	"encoding/json"
	"net/http"
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