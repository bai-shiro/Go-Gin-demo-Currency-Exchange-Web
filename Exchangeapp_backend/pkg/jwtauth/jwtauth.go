package jwtauth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

type Claims struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	TokenType string `json:"token_type"`
	JTI       string `json:"jti"`
	SessionID string `json:"session_id,omitempty"`
	gojwt.RegisteredClaims
}

func GenerateAccessToken(userID uint, username string, secret string, ttl time.Duration) (string, *Claims, error) {
	return generateToken(userID, username, TokenTypeAccess, secret, ttl, "")
}

func GenerateRefreshToken(userID uint, username string, secret string, ttl time.Duration, sessionID string) (string, *Claims, error) {
	if strings.TrimSpace(sessionID) == "" {
		return "", nil, errors.New("missing session_id")
	}
	return generateToken(userID, username, TokenTypeRefresh, secret, ttl, sessionID)
}

func ParseJWT(tokenString, secret string) (*Claims, error) {
	const prefix = "Bearer "
	if !strings.HasPrefix(tokenString, prefix) {
		return nil, errors.New("missing bearer token")
	}

	tokenString = strings.TrimSpace(strings.TrimPrefix(tokenString, prefix))
	return ParseRawJWT(tokenString, secret)
}

func ParseRawJWT(tokenString, secret string) (*Claims, error) {
	tokenString = strings.TrimSpace(tokenString)
	if tokenString == "" {
		return nil, errors.New("missing token")
	}

	token, err := gojwt.ParseWithClaims(tokenString, &Claims{}, func(token *gojwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*gojwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	if claims.UserID == 0 {
		return nil, errors.New("missing user_id claim")
	}
	if claims.TokenType != TokenTypeAccess && claims.TokenType != TokenTypeRefresh {
		return nil, errors.New("invalid token_type claim")
	}
	if strings.TrimSpace(claims.JTI) == "" {
		return nil, errors.New("missing jti claim")
	}
	if claims.TokenType == TokenTypeRefresh && strings.TrimSpace(claims.SessionID) == "" {
		return nil, errors.New("missing session_id claim")
	}

	return claims, nil
}

func NewTokenID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

func generateToken(userID uint, username string, tokenType string, secret string, ttl time.Duration, sessionID string) (string, *Claims, error) {
	jti, err := NewTokenID()
	if err != nil {
		return "", nil, err
	}

	now := time.Now()
	claims := &Claims{
		UserID:    userID,
		Username:  username,
		TokenType: tokenType,
		JTI:       jti,
		SessionID: sessionID,
		RegisteredClaims: gojwt.RegisteredClaims{
			ID:        jti,
			IssuedAt:  gojwt.NewNumericDate(now),
			ExpiresAt: gojwt.NewNumericDate(now.Add(ttl)),
		},
	}

	token := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", nil, err
	}
	return signedToken, claims, nil
}
