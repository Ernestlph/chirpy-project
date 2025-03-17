package main

import (
	"encoding/json"
	"net/http"
)

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		errorResp := errorResponse{
			Error: "Reset endpoint is only available in dev environment",
		}
		jsonResp, _ := json.Marshal(errorResp)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResp)
		return
	}

	err := cfg.dbQueries.DeleteAllUsers(r.Context()) // Delete all users
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorResp := errorResponse{
			Error: "Failed to delete users from database",
		}
		jsonResp, _ := json.Marshal(errorResp)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResp)
		return
	}

	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits and users reset to 0"))
}
