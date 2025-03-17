package main

import (
	"chirpy-project/internal/auth"
	"chirpy-project/internal/database"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

type usersRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (cfg *apiConfig) createUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		errorResp := errorResponse{
			Error: "Method not allowed",
		}
		jsonResp, _ := json.Marshal(errorResp)
		w.Write(jsonResp)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := usersRequest{}
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

	if params.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		errorResp := errorResponse{
			Error: "Email is required",
		}
		jsonResp, _ := json.Marshal(errorResp)
		w.Write(jsonResp)
		return
	}

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

	user, err := cfg.dbQueries.CreateUser(r.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})

	if err != nil && err.Error() == "UNIQUE constraint failed: users.email" {
		w.WriteHeader(http.StatusConflict)
		errorResp := errorResponse{
			Error: "User already exists",
		}
		jsonResp, _ := json.Marshal(errorResp)
		w.Write(jsonResp)
		return
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorResp := errorResponse{
			Error: "Failed to create user",
		}
		jsonResp, _ := json.Marshal(errorResp)
		w.Write(jsonResp)
		return
	}

	resp := User{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	}

	w.WriteHeader(http.StatusCreated)
	jsonResp, _ := json.Marshal(resp)
	w.Write(jsonResp)

}
