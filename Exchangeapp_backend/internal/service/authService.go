package service

import (
	"exchangeapp/internal/apperrors"
	"exchangeapp/internal/models"
	"exchangeapp/internal/repository"
	"exchangeapp/internal/utils"
	"exchangeapp/pkg/jwtauth"
	"time"
)

type UserStore struct {
}

type AuthService struct {
	users     *repository.UserRepository
	jwtSecret string
	jwtTTL    time.Duration
}

func NewAuthService(users *repository.UserRepository, jwtSecret string, jwtTTL time.Duration) *AuthService {
	return &AuthService{users: users, jwtSecret: jwtSecret, jwtTTL: jwtTTL}
}

func (s *AuthService) Register(username, password string) (string, error) {
	hashedPwd, err := utils.HashPassword(password)
	if err != nil {
		return "", err
	}

	user := &models.User{
		Username: username,
		Password: hashedPwd,
	}

	if err := s.users.Create(user); err != nil {
		return "", err
	}

	token, err := jwtauth.GenerateJWT(user.ID, user.Username, s.jwtSecret, s.jwtTTL)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *AuthService) Login(username string, password string) (string, error) {
	user, err := s.users.FindByUsername(username)
	if err != nil {
		return "", apperrors.ErrUnauthorized
	}

	if !utils.CheckPassword(password, user.Password) {
		return "", apperrors.ErrUnauthorized
	}

	token, err := jwtauth.GenerateJWT(user.ID, user.Username, s.jwtSecret, s.jwtTTL)
	if err != nil {
		return "", apperrors.ErrInternal
	}

	return token, nil
}
