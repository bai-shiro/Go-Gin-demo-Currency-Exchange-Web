package jwtauth

import (
	"testing"
	"time"
)

func TestGenerateAndParseJWT(t *testing.T) {
	secret := "test-secret"

	token, _, err := GenerateAccessToken(1, "alice", secret, time.Hour)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
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
	if claims.TokenType != TokenTypeAccess {
		t.Fatalf("TokenType = %q, want access", claims.TokenType)
	}
	if claims.JTI == "" {
		t.Fatal("JTI is empty")
	}
}

func TestParseJWTRejectsWrongSecret(t *testing.T) {
	token, _, err := GenerateAccessToken(1, "alice", "right-secret", time.Hour)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	_, err = ParseJWT("Bearer "+token, "wrong-secret")
	if err == nil {
		t.Fatal("ParseJWT() expected error with wrong secret, got nil")
	}
}

func TestParseJWTRejectsMissingBearerPrefix(t *testing.T) {
	token, _, err := GenerateAccessToken(1, "alice", "test-secret", time.Hour)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	_, err = ParseJWT(token, "test-secret")
	if err == nil {
		t.Fatal("ParseJWT() expected error without Bearer prefix, got nil")
	}
}

func TestParseJWTRejectsExpiredToken(t *testing.T) {
	token, _, err := GenerateAccessToken(1, "alice", "test-secret", -time.Hour)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	_, err = ParseJWT("Bearer "+token, "test-secret")
	if err == nil {
		t.Fatal("ParseJWT() expected expired token error, got nil")
	}
}

func TestParseJWTRejectsMissingUserID(t *testing.T) {
	token, _, err := GenerateAccessToken(0, "alice", "test-secret", time.Hour)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	_, err = ParseJWT("Bearer "+token, "test-secret")
	if err == nil {
		t.Fatal("ParseJWT() expected error with missing user_id, got nil")
	}
}

func TestGenerateAndParseRefreshToken(t *testing.T) {
	token, generatedClaims, err := GenerateRefreshToken(1, "alice", "test-secret", time.Hour, "session-1")
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}

	claims, err := ParseRawJWT(token, "test-secret")
	if err != nil {
		t.Fatalf("ParseRawJWT() error = %v", err)
	}

	if claims.TokenType != TokenTypeRefresh {
		t.Fatalf("TokenType = %q, want refresh", claims.TokenType)
	}
	if claims.SessionID != "session-1" {
		t.Fatalf("SessionID = %q, want session-1", claims.SessionID)
	}
	if claims.JTI != generatedClaims.JTI {
		t.Fatalf("JTI = %q, want %q", claims.JTI, generatedClaims.JTI)
	}
}

func TestGenerateRefreshTokenRejectsMissingSessionID(t *testing.T) {
	_, _, err := GenerateRefreshToken(1, "alice", "test-secret", time.Hour, "")
	if err == nil {
		t.Fatal("GenerateRefreshToken() expected error with missing session_id, got nil")
	}
}
