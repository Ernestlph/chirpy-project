package main

import (
	"chirpy-project/internal/auth"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

type upgradeRequest struct {
	Event string `json:"event"`
	Data  userID `json:"data"`
}

type userID struct {
	User_id uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) upgradeUserHandler(w http.ResponseWriter, r *http.Request) {
	// Check the request header if it has the correct API key, if not return 401 status code
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		errorResp := errorResponse{
			Error: "Unauthorized",
		}
		jsonResp, _ := json.Marshal(errorResp)
		w.Write(jsonResp)
		return
	}

	if apiKey != cfg.polkaKey {
		w.WriteHeader(http.StatusUnauthorized)
		errorResp := errorResponse{
			Error: "Unauthorized",
		}
		fmt.Printf("Correct API Key not found in header: %v\n", err)
		jsonResp, _ := json.Marshal(errorResp)
		w.Write(jsonResp)
		return
	}

	// Check the request body for 'event: user.upgraded' if event is something other than user.upgraded return 204 status code
	decoder := json.NewDecoder(r.Body)
	var params upgradeRequest

	err = decoder.Decode(&params)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errorResp := errorResponse{
			Error: "Invalid request payload",
		}
		jsonResp, _ := json.Marshal(errorResp)
		w.Write(jsonResp)
		return
	}
	requestEvent := params.Event
	requestUserID := params.Data.User_id

	if requestEvent != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		fmt.Printf("Event is not user.upgraded\n")
		return
	}

	// Check if user exists in database, if not return 404
	userRecord, err := cfg.dbQueries.GetUserByID(r.Context(), requestUserID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		errorResp := errorResponse{
			Error: "User not found",
		}
		fmt.Printf("User not found: %v\n", err)
		jsonResp, _ := json.Marshal(errorResp)
		w.Write(jsonResp)
		return
	}

	// Check if user is already chirpy red, if so return 204
	if userRecord.IsChirpyRed {
		w.WriteHeader(http.StatusNoContent)
		fmt.Printf("User is already chirpy red\n")
		return
	}

	// Upgrade user to chirpy red and respond with 204 status code and an empty response body
	err = cfg.dbQueries.UpgradeUser(r.Context(), requestUserID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorResp := errorResponse{
			Error: "Internal Server Error",
		}
		fmt.Printf("Error upgrading user: %v\n", err)
		jsonResp, _ := json.Marshal(errorResp)
		w.Write(jsonResp)
		return
	}
	// Respond with 204 status code and an empty response body
	w.WriteHeader(http.StatusNoContent)
	fmt.Printf("User upgraded successfully\n")
}
