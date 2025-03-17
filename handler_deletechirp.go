package main

import (
	"chirpy-project/internal/auth"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

func (cfg *apiConfig) deleteChirpHandler(w http.ResponseWriter, r *http.Request) {
	// Set the content type to JSON
	w.Header().Set("Content-Type", "application/json")
	// Get the access token from header
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		errorResp := errorResponse{
			Error: "No token found in header",
		}
		fmt.Printf("No token found in header: %v\n", err)
		jsonResp, _ := json.Marshal(errorResp)
		w.Write(jsonResp)
		return
	}
	// Check the access token and get the user ID
	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		errorResp := errorResponse{
			Error: "User not found with refresh token",
		}
		fmt.Printf("User not found with refresh token: %v\n", err)
		jsonResp, _ := json.Marshal(errorResp)
		w.Write(jsonResp)
		return
	}
	// Get the chirp ID from the path
	chirpid := r.PathValue("chirpid")

	// Parse the chirp ID to UUID
	chirpidUUID, err := uuid.Parse(chirpid)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errorResp := errorResponse{
			Error: "Invalid chirp ID",
		}
		jsonResp, _ := json.Marshal(errorResp)
		w.Write(jsonResp)
		return
	}

	// Get the chirprecord from the database, if not found return 404 status code
	chirprecord, err := cfg.dbQueries.GetChirp(r.Context(), chirpidUUID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		errorResp := errorResponse{
			Error: "Chirp not found",
		}
		fmt.Printf("Chirp not found in database: %v\n", err)
		jsonResp, _ := json.Marshal(errorResp)
		w.Write(jsonResp)
		return
	}

	// If chirp does not belong to user, return 403 status code
	if chirprecord.UserID != userID {
		w.WriteHeader(http.StatusForbidden)
		errorResp := errorResponse{
			Error: "Forbidden",
		}
		fmt.Printf("Chirp does not belong to user: %v\n", err)
		fmt.Printf("Chirp belongs to user: %v\n", chirprecord.UserID)
		fmt.Printf("User is: %v\n", userID)
		jsonResp, _ := json.Marshal(errorResp)
		w.Write(jsonResp)
		return
	}

	// Delete the chirp from the database
	err = cfg.dbQueries.DeleteChirp(r.Context(), chirpidUUID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorResp := errorResponse{
			Error: "Internal Server Error",
		}
		fmt.Printf("Error deleting chirp: %v\n", err)
		jsonResp, _ := json.Marshal(errorResp)
		w.Write(jsonResp)
		return
	}

	// If chirp deleted successfully return 204 status code
	w.WriteHeader(http.StatusNoContent)
}
