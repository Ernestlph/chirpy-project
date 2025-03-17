package auth_test

import (
	"chirpy-project/internal/auth" // Replace "chirpy-project" with your actual module path
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "testsecret"
	expiresIn := time.Hour

	tokenString, err := auth.MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT failed: %v", err)
	}

	if tokenString == "" {
		t.Errorf("MakeJWT should return a non-empty token string")
	}

	// Basic check that the token string looks like a JWT (has 3 parts separated by dots)
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		t.Errorf("Token string does not have the right format for a JWT, expected 3 parts, got %d", len(parts))
	}
}

func TestValidateJWTSucceeds(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "testsecret"
	expiresIn := time.Hour

	tokenString, err := auth.MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT failed: %v", err)
	}

	validatedUserID, err := auth.ValidateJWT(tokenString, tokenSecret)
	if err != nil {
		t.Fatalf("ValidateJWT failed: %v", err)
	}

	if validatedUserID != userID {
		t.Errorf("ValidateJWT returned incorrect user ID. Expected %v, got %v", userID, validatedUserID)
	}
}

func TestValidateJWTExpired(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "testsecret"
	expiresIn := -time.Hour // Token expired in the past

	tokenString, err := auth.MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT failed: %v", err)
	}

	_, err = auth.ValidateJWT(tokenString, tokenSecret)
	if err == nil {
		t.Errorf("ValidateJWT should have failed for an expired token")
	}

	// Since the error is wrapped, check the error message instead
	if !strings.Contains(err.Error(), "token is expired") &&
		!strings.Contains(err.Error(), "token has expired") {
		t.Errorf("Expected expired token error message, got: %v", err)
	}
}

func TestValidateJWTWrongSecret(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "testsecret"
	wrongSecret := "wrongsecret"
	expiresIn := time.Hour

	tokenString, err := auth.MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT failed: %v", err)
	}

	_, err = auth.ValidateJWT(tokenString, wrongSecret)
	if err == nil {
		t.Errorf("ValidateJWT should have failed for a token with the wrong secret")
	}

	// Since the error is wrapped, check the error message instead
	if !strings.Contains(err.Error(), "signature is invalid") &&
		!strings.Contains(err.Error(), "invalid signature") {
		t.Errorf("Expected signature validation error message, got: %v", err)
	}
}

func TestValidateJWTInvalidToken(t *testing.T) {
	tokenSecret := "testsecret"
	invalidTokenString := "this.is.not.a.valid.jwt" // A deliberately invalid token

	_, err := auth.ValidateJWT(invalidTokenString, tokenSecret)
	if err == nil {
		t.Errorf("ValidateJWT should have failed for an invalid token format")
	}
}
