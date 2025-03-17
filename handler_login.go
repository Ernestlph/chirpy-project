package main

import (
	"chirpy-project/internal/auth"
	"chirpy-project/internal/database"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type loginRequest struct {
	Password string `json:"password"`
	Email    string `json:"email"`
	Expires  int    `json:"expires_in_seconds"`
}

type loginResponse struct {
	ID            uuid.UUID `json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Email         string    `json:"email"`
	Token         string    `json:"token"`
	Refresh_Token string    `json:"refresh_token"`
	IsChirpyRed   bool      `json:"is_chirpy_red"`
}

func (cfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the request method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	params := loginRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errorResp := errorResponse{
			Error: "Invalid request payload",
		}
		jsonResp, _ := json.Marshal(errorResp)
		w.Write(jsonResp)
		return
	}

	// Extract the username and password and expires time from params
	username := params.Email
	password := params.Password

	// expires must always be 1 hour
	params.Expires = 3600
	expires := 3600

	// Based on email, check if the user exists in the database
	user, err := cfg.dbQueries.Login(r.Context(), username)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		errorResp := errorResponse{
			Error: "Incorrect email or password",
		}
		jsonResp, _ := json.Marshal(errorResp)
		w.Write(jsonResp)
		return
	}

	// Check if the password is correct
	if err := auth.CheckPasswordHash(password, user.HashedPassword); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		errorResp := errorResponse{
			Error: "Incorrect email or password",
		}
		jsonResp, _ := json.Marshal(errorResp)
		w.Write(jsonResp)
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Duration(expires)*time.Second)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorResp := errorResponse{
			Error: "Failed to generate token",
		}
		jsonResp, _ := json.Marshal(errorResp)
		w.Write(jsonResp)
		return
	}

	// Generatea refresh token from auth
	refreshtoken, err := auth.MakeRefreshToken()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorResp := errorResponse{
			Error: "Failed to generate refresh token",
		}
		jsonResp, _ := json.Marshal(errorResp)
		w.Write(jsonResp)
		return
	}

	// Insert the refresh token into the database
	_, err = cfg.dbQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		UserID:    user.ID,
		Token:     refreshtoken,
		ExpiresAt: time.Now().Add(time.Duration(1440) * time.Hour),
		RevokedAt: sql.NullTime{},
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorResp := errorResponse{
			Error: "Failed to create refresh token",
		}
		jsonResp, _ := json.Marshal(errorResp)
		w.Write(jsonResp)
		return
	}

	resp := loginResponse{
		ID:            user.ID,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
		Email:         user.Email,
		Token:         token,
		Refresh_Token: refreshtoken,
		IsChirpyRed:   user.IsChirpyRed,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	jsonResp, _ := json.Marshal(resp)
	w.Write(jsonResp)

}
