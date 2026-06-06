package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"exchangeapp/internal/apperrors"
	"exchangeapp/internal/models"
	"exchangeapp/pkg/jwtauth"
	"exchangeapp/pkg/passwordbcrypt"
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

type userStore interface {
	Create(user *models.User) error
	FindByUsername(username string) (*models.User, error)
}

type AuthService struct {
	users              userStore
	redis              *redis.Client
	jwtSecret          string
	accessTTL          time.Duration
	refreshSlidingTTL  time.Duration
	refreshAbsoluteTTL time.Duration
}

type JWTOptions struct {
	Secret             string
	AccessTTL          time.Duration
	RefreshSlidingTTL  time.Duration
	RefreshAbsoluteTTL time.Duration
}

type AuthTokens struct {
	AccessToken           string
	RefreshToken          string
	AccessTokenExpiresIn  int64
	RefreshTokenExpiresIn int64
}

type refreshTokenState struct {
	UserID    uint      `json:"userID"`
	Username  string    `json:"username"`
	SessionID string    `json:"sessionID"`
	TokenHash string    `json:"tokenHash"`
	CreatedAt time.Time `json:"createdAt"`
}

type refreshSessionState struct {
	UserID            uint      `json:"userID"`
	Username          string    `json:"username"`
	CurrentJTI        string    `json:"currentJTI"`
	AbsoluteExpiresAt time.Time `json:"absoluteExpiresAt"`
	CreatedAt         time.Time `json:"createdAt"`
}

func NewAuthService(users userStore, redisClient *redis.Client, jwtOptions JWTOptions) *AuthService {
	return &AuthService{
		users:              users,
		redis:              redisClient,
		jwtSecret:          jwtOptions.Secret,
		accessTTL:          jwtOptions.AccessTTL,
		refreshSlidingTTL:  jwtOptions.RefreshSlidingTTL,
		refreshAbsoluteTTL: jwtOptions.RefreshAbsoluteTTL,
	}
}

func (s *AuthService) Register(username, password string) (*AuthTokens, error) {
	hashedPwd, err := passwordbcrypt.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Username: username,
		Password: hashedPwd,
	}

	if err := s.users.Create(user); err != nil {
		return nil, err
	}

	return s.issueTokenPair(user.ID, user.Username, "", time.Now())
}

func (s *AuthService) Login(username string, password string) (*AuthTokens, error) {
	user, err := s.users.FindByUsername(username)
	if err != nil {
		return nil, apperrors.ErrUnauthorized
	}

	if !passwordbcrypt.CheckPassword(password, user.Password) {
		return nil, apperrors.ErrUnauthorized
	}

	return s.issueTokenPair(user.ID, user.Username, "", time.Now())
}

func (s *AuthService) Refresh(refreshToken string) (*AuthTokens, error) {
	claims, err := jwtauth.ParseRawJWT(refreshToken, s.jwtSecret)
	if err != nil || claims.TokenType != jwtauth.TokenTypeRefresh {
		return nil, apperrors.ErrUnauthorized
	}

	tokenState, err := s.loadRefreshTokenState(claims.JTI)
	if err != nil {
		s.revokeRefreshSession(claims.UserID, claims.SessionID, "")
		return nil, apperrors.ErrUnauthorized
	}
	if tokenState.UserID != claims.UserID ||
		tokenState.Username != claims.Username ||
		tokenState.SessionID != claims.SessionID ||
		tokenState.TokenHash != hashToken(refreshToken) {
		s.revokeRefreshSession(claims.UserID, claims.SessionID, claims.JTI)
		return nil, apperrors.ErrUnauthorized
	}

	sessionState, err := s.loadRefreshSessionState(claims.SessionID)
	if err != nil {
		s.revokeRefreshSession(claims.UserID, claims.SessionID, claims.JTI)
		return nil, apperrors.ErrUnauthorized
	}
	if sessionState.UserID != claims.UserID ||
		sessionState.Username != claims.Username ||
		sessionState.CurrentJTI != claims.JTI {
		s.revokeRefreshSession(claims.UserID, claims.SessionID, claims.JTI)
		return nil, apperrors.ErrUnauthorized
	}

	now := time.Now()
	if !now.Before(sessionState.AbsoluteExpiresAt) {
		s.revokeRefreshSession(claims.UserID, claims.SessionID, claims.JTI)
		return nil, apperrors.ErrUnauthorized
	}

	if err := s.redis.Del(refreshTokenKey(claims.JTI)).Err(); err != nil {
		return nil, apperrors.ErrInternal
	}

	tokens, err := s.issueTokenPair(claims.UserID, claims.Username, claims.SessionID, now)
	if err != nil {
		return nil, apperrors.ErrInternal
	}
	return tokens, nil
}

func (s *AuthService) Logout(refreshToken string) error {
	claims, err := jwtauth.ParseRawJWT(refreshToken, s.jwtSecret)
	if err != nil || claims.TokenType != jwtauth.TokenTypeRefresh {
		return apperrors.ErrUnauthorized
	}

	s.revokeRefreshSession(claims.UserID, claims.SessionID, claims.JTI)
	return nil
}

func (s *AuthService) issueTokenPair(userID uint, username string, sessionID string, now time.Time) (*AuthTokens, error) {
	if s.redis == nil {
		return nil, fmt.Errorf("redis client is nil")
	}
	if sessionID == "" {
		newSessionID, err := jwtauth.NewTokenID()
		if err != nil {
			return nil, err
		}
		sessionID = newSessionID
	}

	absoluteExpiresAt := now.Add(s.refreshAbsoluteTTL)
	existingSession, err := s.loadRefreshSessionState(sessionID)
	if err == nil {
		absoluteExpiresAt = existingSession.AbsoluteExpiresAt
	} else if err != redis.Nil {
		return nil, err
	}

	refreshTTL := minDuration(s.refreshSlidingTTL, time.Until(absoluteExpiresAt))
	if refreshTTL <= 0 {
		return nil, apperrors.ErrUnauthorized
	}

	accessToken, _, err := jwtauth.GenerateAccessToken(userID, username, s.jwtSecret, s.accessTTL)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshClaims, err := jwtauth.GenerateRefreshToken(userID, username, s.jwtSecret, refreshTTL, sessionID)
	if err != nil {
		return nil, err
	}

	if err := s.storeRefreshToken(refreshToken, refreshClaims, refreshTTL); err != nil {
		return nil, err
	}
	if err := s.storeRefreshSession(userID, username, sessionID, refreshClaims.JTI, absoluteExpiresAt, now); err != nil {
		return nil, err
	}
	if err := s.redis.SAdd(refreshUserSessionsKey(userID), sessionID).Err(); err != nil {
		return nil, err
	}
	if err := s.redis.Expire(refreshUserSessionsKey(userID), time.Until(absoluteExpiresAt)).Err(); err != nil {
		return nil, err
	}

	return &AuthTokens{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresIn:  int64(s.accessTTL.Seconds()),
		RefreshTokenExpiresIn: int64(refreshTTL.Seconds()),
	}, nil
}

func (s *AuthService) storeRefreshToken(refreshToken string, claims *jwtauth.Claims, ttl time.Duration) error {
	state := refreshTokenState{
		UserID:    claims.UserID,
		Username:  claims.Username,
		SessionID: claims.SessionID,
		TokenHash: hashToken(refreshToken),
		CreatedAt: time.Now(),
	}
	payload, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return s.redis.Set(refreshTokenKey(claims.JTI), payload, ttl).Err()
}

func (s *AuthService) storeRefreshSession(userID uint, username string, sessionID string, currentJTI string, absoluteExpiresAt time.Time, now time.Time) error {
	ttl := time.Until(absoluteExpiresAt)
	if ttl <= 0 {
		return apperrors.ErrUnauthorized
	}

	state := refreshSessionState{
		UserID:            userID,
		Username:          username,
		CurrentJTI:        currentJTI,
		AbsoluteExpiresAt: absoluteExpiresAt,
		CreatedAt:         now,
	}
	payload, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return s.redis.Set(refreshSessionKey(sessionID), payload, ttl).Err()
}

func (s *AuthService) loadRefreshTokenState(jti string) (*refreshTokenState, error) {
	data, err := s.redis.Get(refreshTokenKey(jti)).Result()
	if err != nil {
		return nil, err
	}
	var state refreshTokenState
	if err := json.Unmarshal([]byte(data), &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func (s *AuthService) loadRefreshSessionState(sessionID string) (*refreshSessionState, error) {
	data, err := s.redis.Get(refreshSessionKey(sessionID)).Result()
	if err != nil {
		return nil, err
	}
	var state refreshSessionState
	if err := json.Unmarshal([]byte(data), &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func (s *AuthService) revokeRefreshSession(userID uint, sessionID string, jti string) {
	if s.redis == nil {
		return
	}
	if jti == "" && sessionID != "" {
		if sessionState, err := s.loadRefreshSessionState(sessionID); err == nil {
			jti = sessionState.CurrentJTI
		}
	}
	if jti != "" {
		_ = s.redis.Del(refreshTokenKey(jti)).Err()
	}
	if sessionID != "" {
		_ = s.redis.Del(refreshSessionKey(sessionID)).Err()
	}
	if userID != 0 && sessionID != "" {
		_ = s.redis.SRem(refreshUserSessionsKey(userID), sessionID).Err()
	}
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func refreshTokenKey(jti string) string {
	return fmt.Sprintf("auth:refresh:token:%s", jti)
}

func refreshSessionKey(sessionID string) string {
	return fmt.Sprintf("auth:refresh:session:%s", sessionID)
}

func refreshUserSessionsKey(userID uint) string {
	return fmt.Sprintf("auth:refresh:user:%d", userID)
}

func minDuration(a time.Duration, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
