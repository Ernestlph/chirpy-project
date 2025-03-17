package main

import (
	"chirpy-project/internal/auth"
	"chirpy-project/internal/database"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type updateUserRequest struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

type updateUserResponse struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

func (cfg *apiConfig) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Extract the access token from the Authorization header
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		errorResp := errorResponse{
			Error: "Unauthorized",
		}
		jsonResp, _ := json.Marshal(errorResp)
		w.Write(jsonResp)
		return
	}

	// Validate the access token and get the user ID
	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		errorResp := errorResponse{
			Error: "Unauthorized",
		}
		jsonResp, _ := json.Marshal(errorResp)
		w.Write(jsonResp)
		return
	}

	// Decode the request body
	decoder := json.NewDecoder(r.Body)
	params := updateUserRequest{}
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

	// Hash the new password
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorResp := errorResponse{
			Error: "Failed to hash password",
		}
		jsonResp, _ := json.Marshal(errorResp)
		w.Write(jsonResp)
		return
	}

	// Update the user in the database
	updatedUser, err := cfg.dbQueries.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:             userID,
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorResp := errorResponse{
			Error: "Failed to update user",
		}
		jsonResp, _ := json.Marshal(errorResp)
		w.Write(jsonResp)
		return
	}

	// Respond with the updated user information (excluding password)
	response := updateUserResponse{
		ID:          updatedUser.ID,
		CreatedAt:   updatedUser.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   updatedUser.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		Email:       updatedUser.Email,
		IsChirpyRed: updatedUser.IsChirpyRed,
	}

	w.WriteHeader(http.StatusOK)
	jsonResp, _ := json.Marshal(response)
	w.Write(jsonResp)
}
