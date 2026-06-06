package service

import (
	"exchangeapp/internal/apperrors"
	"exchangeapp/internal/models"
	"exchangeapp/pkg/jwtauth"
	"exchangeapp/pkg/passwordbcrypt"
	"testing"
	"time"

	"github.com/go-redis/redis"
	"gorm.io/gorm"
)

type fakeUserStore struct {
	user *models.User
}

func (f *fakeUserStore) Create(user *models.User) error {
	user.ID = 1
	f.user = user
	return nil
}

func (f *fakeUserStore) FindByUsername(username string) (*models.User, error) {
	if f.user == nil || f.user.Username != username {
		return nil, apperrors.ErrUnauthorized
	}
	return f.user, nil
}

func TestAuthServiceLoginReturnsTokenPair(t *testing.T) {
	rdb := newTestRedis(t)
	users := newFakeUserStore(t, 1, "alice", "secret")
	svc := NewAuthService(users, rdb, testJWTOptions(time.Minute, time.Hour, 24*time.Hour))

	tokens, err := svc.Login("alice", "secret")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatalf("Login() tokens = %+v, want access and refresh token", tokens)
	}

	accessClaims, err := jwtauth.ParseJWT("Bearer "+tokens.AccessToken, "test-secret")
	if err != nil {
		t.Fatalf("parse access token error = %v", err)
	}
	if accessClaims.TokenType != jwtauth.TokenTypeAccess {
		t.Fatalf("access token type = %q, want access", accessClaims.TokenType)
	}

	refreshClaims, err := jwtauth.ParseRawJWT(tokens.RefreshToken, "test-secret")
	if err != nil {
		t.Fatalf("parse refresh token error = %v", err)
	}
	if refreshClaims.TokenType != jwtauth.TokenTypeRefresh {
		t.Fatalf("refresh token type = %q, want refresh", refreshClaims.TokenType)
	}
	if refreshClaims.SessionID == "" {
		t.Fatal("refresh token session_id is empty")
	}
	if exists, err := refreshTokenExists(rdb, refreshClaims.JTI); err != nil || !exists {
		t.Fatalf("refresh token redis state exists = %v, err = %v; want true, nil", exists, err)
	}
}

func TestAuthServiceRefreshRotatesRefreshToken(t *testing.T) {
	rdb := newTestRedis(t)
	users := newFakeUserStore(t, 1, "alice", "secret")
	svc := NewAuthService(users, rdb, testJWTOptions(time.Minute, time.Hour, 24*time.Hour))

	tokens, err := svc.Login("alice", "secret")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	oldClaims, err := jwtauth.ParseRawJWT(tokens.RefreshToken, "test-secret")
	if err != nil {
		t.Fatalf("parse old refresh token error = %v", err)
	}

	rotated, err := svc.Refresh(tokens.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	newClaims, err := jwtauth.ParseRawJWT(rotated.RefreshToken, "test-secret")
	if err != nil {
		t.Fatalf("parse new refresh token error = %v", err)
	}

	if newClaims.JTI == oldClaims.JTI {
		t.Fatal("Refresh() reused old refresh token jti")
	}
	if newClaims.SessionID != oldClaims.SessionID {
		t.Fatalf("new session_id = %q, want %q", newClaims.SessionID, oldClaims.SessionID)
	}
	if exists, err := refreshTokenExists(rdb, oldClaims.JTI); err != nil || exists {
		t.Fatalf("old refresh token redis state exists = %v, err = %v; want false, nil", exists, err)
	}

	_, err = svc.Refresh(tokens.RefreshToken)
	if err != apperrors.ErrUnauthorized {
		t.Fatalf("Refresh(old token) error = %v, want ErrUnauthorized", err)
	}

	_, err = svc.Refresh(rotated.RefreshToken)
	if err != apperrors.ErrUnauthorized {
		t.Fatalf("Refresh(rotated token after reuse detection) error = %v, want ErrUnauthorized", err)
	}
}

func TestAuthServiceRefreshRejectsAccessToken(t *testing.T) {
	rdb := newTestRedis(t)
	users := newFakeUserStore(t, 1, "alice", "secret")
	svc := NewAuthService(users, rdb, testJWTOptions(time.Minute, time.Hour, 24*time.Hour))

	tokens, err := svc.Login("alice", "secret")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	_, err = svc.Refresh(tokens.AccessToken)
	if err != apperrors.ErrUnauthorized {
		t.Fatalf("Refresh(access token) error = %v, want ErrUnauthorized", err)
	}
}

func TestAuthServiceLogoutRevokesRefreshSession(t *testing.T) {
	rdb := newTestRedis(t)
	users := newFakeUserStore(t, 1, "alice", "secret")
	svc := NewAuthService(users, rdb, testJWTOptions(time.Minute, time.Hour, 24*time.Hour))

	tokens, err := svc.Login("alice", "secret")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if err := svc.Logout(tokens.RefreshToken); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}

	_, err = svc.Refresh(tokens.RefreshToken)
	if err != apperrors.ErrUnauthorized {
		t.Fatalf("Refresh(logged out token) error = %v, want ErrUnauthorized", err)
	}
}

func TestAuthServiceRefreshTTLIsCappedByAbsoluteExpiration(t *testing.T) {
	rdb := newTestRedis(t)
	users := newFakeUserStore(t, 1, "alice", "secret")
	svc := NewAuthService(users, rdb, testJWTOptions(time.Minute, 24*time.Hour, 5*time.Second))

	tokens, err := svc.Login("alice", "secret")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if tokens.RefreshTokenExpiresIn <= 0 || tokens.RefreshTokenExpiresIn > 5 {
		t.Fatalf("RefreshTokenExpiresIn = %d, want in range 1..5", tokens.RefreshTokenExpiresIn)
	}
}

func newFakeUserStore(t *testing.T, userID uint, username string, password string) *fakeUserStore {
	t.Helper()
	hashedPassword, err := passwordbcrypt.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	return &fakeUserStore{
		user: &models.User{
			Model:    gorm.Model{ID: userID},
			Username: username,
			Password: hashedPassword,
		},
	}
}

func testJWTOptions(accessTTL time.Duration, refreshSlidingTTL time.Duration, refreshAbsoluteTTL time.Duration) JWTOptions {
	return JWTOptions{
		Secret:             "test-secret",
		AccessTTL:          accessTTL,
		RefreshSlidingTTL:  refreshSlidingTTL,
		RefreshAbsoluteTTL: refreshAbsoluteTTL,
	}
}

func refreshTokenExists(rdb *redis.Client, jti string) (bool, error) {
	count, err := rdb.Exists(refreshTokenKey(jti)).Result()
	return count == 1, err
}
