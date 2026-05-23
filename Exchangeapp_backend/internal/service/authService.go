package service

import (
	"exchangeapp/internal/apperrors"
	"exchangeapp/internal/models"
	"exchangeapp/internal/repository"
	"exchangeapp/internal/utils"
)

type UserStore struct {
}

type AuthService struct {
	users *repository.UserRepository
}

func NewAuthService(users *repository.UserRepository) *AuthService {
	return &AuthService{users: users}
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

	token, err := utils.GenerateJWT(user.Username)
	if err != nil {
		return "", err
	}

	return token, s.users.Create(user)
}

func (s *AuthService) Login(username string, password string) (string, error) {
	user, err := s.users.FindByUsername(username)
	if err != nil {
		return "", apperrors.ErrUnauthorized
	}

	if !utils.CheckPassword(password, user.Password) {
		return "", apperrors.ErrUnauthorized
	}

	token, err := utils.GenerateJWT(user.Username)
	if err != nil {
		return "", apperrors.ErrInternal
	}

	return token, nil
}