package main

import (
	"chirpy-project/internal/auth"
	"encoding/json"
	"fmt"
	"net/http"
)

func (cfg *apiConfig) revokeRefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		errorResp := errorResponse{
			Error: "Unauthorized",
		}
		fmt.Printf("Token not found in header: %v\n", err)
		jsonResp, _ := json.Marshal(errorResp)
		w.Write(jsonResp)
		return
	}

	refreshTokenRecord, err := cfg.dbQueries.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		errorResp := errorResponse{
			Error: "Refresh token not found",
		}
		fmt.Printf("Refresh Token not found in database: %v\n", err)
		jsonResp, _ := json.Marshal(errorResp)
		w.Write(jsonResp)
		return
	}
	err = cfg.dbQueries.RevokeRefreshToken(r.Context(), refreshTokenRecord.Token)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorResp := errorResponse{
			Error: "Internal Server Error",
		}
		fmt.Printf("Error revoking refresh token: %v\n", err)
		jsonResp, _ := json.Marshal(errorResp)
		w.Write(jsonResp)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNoContent)
	w.Write([]byte("Refresh token revoked"))
}
