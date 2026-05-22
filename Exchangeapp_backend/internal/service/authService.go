package service

import (
	"errors"
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

func (s *AuthService) Register(user *models.User) (string, error) {
	hashedPwd, err := utils.HashPassword(user.Password)
	if err != nil {
		return "", err
	}

	user.Password = hashedPwd

	token, err := utils.GenerateJWT(user.Username)
	if err != nil {
		return "", err
	}

	return token, s.users.Create(user)
}

func (s *AuthService) Login(username string, password string) (string, error) {
	user, err := s.users.FindByUsername(username)
	if err != nil {
		return "", errors.New("incorrect credentials")
	}

	if !utils.CheckPassword(password, user.Password) {
		return "", errors.New("incorrect credentials")
	}

	token, err := utils.GenerateJWT(user.Username)
	if err != nil {
		return "", errors.New("incorrect credentials")
	}

	return token, nil
}