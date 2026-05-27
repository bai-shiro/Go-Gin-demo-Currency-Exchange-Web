package jwtauth

import (
	"errors"
	"strings"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	gojwt.RegisteredClaims
}

func GenerateJWT(userID uint, username string, secret string, ttl time.Duration) (string, error) {
	now := time.Now()
	token := gojwt.NewWithClaims(gojwt.SigningMethodHS256, Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: gojwt.RegisteredClaims{
			IssuedAt:  gojwt.NewNumericDate(now),
			ExpiresAt: gojwt.NewNumericDate(now.Add(ttl)),
		},
	})

	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func ParseJWT(tokenString, secret string) (*Claims, error) {
	const prefix = "Bearer "
	if !strings.HasPrefix(tokenString, prefix) {
		return nil, errors.New("missing bearer token")
	}

	tokenString = strings.TrimSpace(strings.TrimPrefix(tokenString, prefix))

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

	return claims, nil
}
