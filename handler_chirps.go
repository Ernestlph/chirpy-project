package main

import (
	"chirpy-project/internal/auth"
	"chirpy-project/internal/database"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

type chirpRequest struct {
	Body   string `json:"body"`
	UserID string `json:"user_id"`
}

type chirpResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

type errorResponse struct {
	Error string `json:"error"`
}

var profaneWords = []string{"kerfuffle", "sharbert", "fornax"}

func (cfg *apiConfig) createChirpHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	defer r.Body.Close()

	fmt.Printf("Request method: %s\n", r.Method)
	fmt.Printf("Authorization header: %s\n", r.Header.Get("Authorization"))

	// Parse the JWT from the Authorization header
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		fmt.Printf("GetBearerToken error: %v\n", err)
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Validate the JWT
	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		fmt.Printf("ValidateJWT error: %v\n", err)
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := chirpRequest{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate chirp length
	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	// Filter profane words - improved case handling
	cleanedBody := params.Body
	for _, word := range profaneWords {
		wordLower := strings.ToLower(word)
		re := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(wordLower))
		cleanedBody = re.ReplaceAllString(cleanedBody, "****")
	}

	// Create the chirp in the database, using the userID from the JWT
	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleanedBody,
		UserID: userID, // Use userID from JWT
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create chirp")
		return
	}

	// Create the response
	resp := chirpResponse{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}

	respondWithJSON(w, http.StatusCreated, resp)
}

// Helper function to respond with an error
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, errorResponse{Error: message})
}

// Helper function to respond with JSON
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func (cfg *apiConfig) listChirpsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sort_desc := false
	// Check if sorting decision is in the query parameters in the URL
	sort_Type := r.URL.Query().Get("sort")
	if sort_Type == "desc" {
		sort_desc = true
	}
	// Check if author_id is in the query parameters in the URL
	author_id := r.URL.Query().Get("author_id")
	if author_id != "" {
		parsed_author_id, err := uuid.Parse(author_id)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			errorResp := errorResponse{
				Error: "Invalid author_id",
			}
			jsonResp, _ := json.Marshal(errorResp)
			w.Write(jsonResp)
			return
		}
		chirpResponses, err := cfg.dbQueries.ListChirpsByUser(r.Context(), parsed_author_id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			errorResp := errorResponse{
				Error: "Failed to list chirps from database",
			}
			jsonResp, _ := json.Marshal(errorResp)
			w.Write(jsonResp)
			return
		}
		chirps := []chirpResponse{}
		for _, chirpRecord := range chirpResponses {
			chirps = append(chirps, chirpResponse{
				ID:        chirpRecord.ID,
				CreatedAt: chirpRecord.CreatedAt,
				UpdatedAt: chirpRecord.UpdatedAt,
				Body:      chirpRecord.Body,
				UserID:    chirpRecord.UserID,
			})
		}
		if sort_desc {
			sort.Slice(chirps, func(i, j int) bool {
				return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
			})
		}
		w.WriteHeader(http.StatusOK)
		jsonResp, _ := json.Marshal(chirps)
		w.Write(jsonResp)
		return

	}

	chirpsDB, err := cfg.dbQueries.ListChirps(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorResp := errorResponse{
			Error: "Failed to list chirps from database",
		}
		jsonResp, _ := json.Marshal(errorResp)
		w.Write(jsonResp)
		return
	}

	chirps := []chirpResponse{}
	for _, chirpDB := range chirpsDB {
		chirps = append(chirps, chirpResponse{
			ID:        chirpDB.ID,
			CreatedAt: chirpDB.CreatedAt,
			UpdatedAt: chirpDB.UpdatedAt,
			Body:      chirpDB.Body,
			UserID:    chirpDB.UserID,
		})
	}

	if sort_desc {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
		})
	}
	w.WriteHeader(http.StatusOK)
	jsonResp, _ := json.Marshal(chirps)
	w.Write(jsonResp)
}

func (cfg *apiConfig) getChirpHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		errorResp := errorResponse{
			Error: "Method not allowed",
		}
		jsonResp, _ := json.Marshal(errorResp)
		w.Write(jsonResp)
		return
	}

	id := r.PathValue("chirpid")
	idUUID, err := uuid.Parse(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errorResp := errorResponse{
			Error: "Invalid chirp ID",
		}
		jsonResp, _ := json.Marshal(errorResp)
		w.Write(jsonResp)
		return
	}

	chirpDB, err := cfg.dbQueries.GetChirp(r.Context(), idUUID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		errorResp := errorResponse{
			Error: "Chirp not found",
		}
		jsonResp, _ := json.Marshal(errorResp)
		w.Write(jsonResp)
		return
	}

	chirp := chirpResponse{
		ID:        chirpDB.ID,
		CreatedAt: chirpDB.CreatedAt,
		UpdatedAt: chirpDB.UpdatedAt,
		Body:      chirpDB.Body,
		UserID:    chirpDB.UserID,
	}

	w.WriteHeader(http.StatusOK)
	jsonResp, _ := json.Marshal(chirp)
	w.Write(jsonResp)
}
