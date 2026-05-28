package jwtauth

import (
	"testing"
	"time"
)

func TestGenerateAndParseJWT(t *testing.T) {
	secret := "test-secret"

	token, err := GenerateJWT(1, "alice", secret, time.Hour)
	if err != nil {
		t.Fatalf("GenerateJWT() error = %v", err)
	}

	claims, err := ParseJWT("Bearer "+token, secret)
	if err != nil {
		t.Fatalf("ParseJWT() error = %v", err)
	}

	if claims.UserID != 1 {
		t.Fatalf("UserID = %d, want 1", claims.UserID)
	}

	if claims.Username != "alice" {
		t.Fatalf("Username = %q, want alice", claims.Username)
	}
}

func TestParseJWTRejectsWrongSecret(t *testing.T) {
	token, err := GenerateJWT(1, "alice", "right-secret", time.Hour)
	if err != nil {
		t.Fatalf("GenerateJWT() error = %v", err)
	}

	_, err = ParseJWT("Bearer "+token, "wrong-secret")
	if err == nil {
		t.Fatal("ParseJWT() expected error with wrong secret, got nil")
	}
}

func TestParseJWTRejectsMissingBearerPrefix(t *testing.T) {
	token, err := GenerateJWT(1, "alice", "test-secret", time.Hour)
	if err != nil {
		t.Fatalf("GenerateJWT() error = %v", err)
	}

	_, err = ParseJWT(token, "test-secret")
	if err == nil {
		t.Fatal("ParseJWT() expected error without Bearer prefix, got nil")
	}
}

func TestParseJWTRejectsExpiredToken(t *testing.T) {
	token, err := GenerateJWT(1, "alice", "test-secret", -time.Hour)
	if err != nil {
		t.Fatalf("GenerateJWT() error = %v", err)
	}

	_, err = ParseJWT("Bearer "+token, "test-secret")
	if err == nil {
		t.Fatal("ParseJWT() expected expired token error, got nil")
	}
}

func TestParseJWTRejectsMissingUserID(t *testing.T) {
	token, err := GenerateJWT(0, "alice", "test-secret", time.Hour)
	if err != nil {
		t.Fatalf("GenerateJWT() error = %v", err)
	}

	_, err = ParseJWT("Bearer "+token, "test-secret")
	if err == nil {
		t.Fatal("ParseJWT() expected error with missing user_id, got nil")
	}
}
