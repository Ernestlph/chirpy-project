package main

import (
	"chirpy-project/internal/auth"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type accessTokenResponse struct {
	Token string `json:"token"`
}

func (cfg *apiConfig) refreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	requesttoken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		errorResp := errorResponse{
			Error: "Unauthorized",
		}
		jsonResp, _ := json.Marshal(errorResp)
		w.Write(jsonResp)
		return
	}
	if requesttoken == "" {
		w.WriteHeader(http.StatusUnauthorized)
		errorResp := errorResponse{
			Error: "Unauthorized",
		}
		fmt.Printf("GetBearerToken not found: %v\n", err)
		jsonResp, _ := json.Marshal(errorResp)
		w.Write(jsonResp)
		return
	}
	user, err := cfg.dbQueries.GetUserFromRefreshToken(r.Context(), requesttoken)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		errorResp := errorResponse{
			Error: "Unauthorized",
		}
		fmt.Printf("User not found with refresh token: %v\n", err)
		jsonResp, _ := json.Marshal(errorResp)
		w.Write(jsonResp)
		return
	}

	if user.ExpiresAt.Before(time.Now()) {
		w.WriteHeader(http.StatusUnauthorized)
		errorResp := errorResponse{
			Error: "Unauthorized",
		}
		fmt.Printf("User refresh token has expired: %v\n", err)
		jsonResp, _ := json.Marshal(errorResp)
		w.Write(jsonResp)
		return
	}

	// Check if the refresh token has been revoked
	if user.RevokedAt.Valid {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		errorResp := errorResponse{
			Error: "Unauthorized - token revoked",
		}
		fmt.Printf("User refresh token has been revoked at: %v\n", user.RevokedAt.Time)
		jsonResp, _ := json.Marshal(errorResp)
		w.Write(jsonResp)
		return
	}

	// User refresh token is valid, generate a new access token
	var accessToken string
	accessToken, err = auth.MakeJWT(user.UserID, cfg.jwtSecret, time.Hour)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorResp := errorResponse{
			Error: "Failed to generate access token",
		}
		jsonResp, _ := json.Marshal(errorResp)
		w.Write(jsonResp)
		return
	}

	refreshToken := accessTokenResponse{
		Token: accessToken,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	jsonResp, _ := json.Marshal(refreshToken)
	w.Write(jsonResp)
}
