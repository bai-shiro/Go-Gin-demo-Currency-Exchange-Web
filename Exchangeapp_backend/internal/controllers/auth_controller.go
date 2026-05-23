package controllers

import (
	"exchangeapp/internal/apperrors"
	"exchangeapp/internal/dto"
	"exchangeapp/internal/response"
	"exchangeapp/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	auth *service.AuthService
}

func NewAuthController(auth *service.AuthService) *AuthController {
	return &AuthController{auth: auth}
}

func (c *AuthController) Register(ctx *gin.Context) {
	var req dto.RegisterRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, apperrors.ErrInvalidParams)
		return
	}

	token, err := c.auth.Register(req.Username, req.Password)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, dto.AuthResponse{Token: token})
}

func (c *AuthController) Login(ctx *gin.Context) {
	var req dto.LoginRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, apperrors.ErrInvalidParams)
		return
	}

	token, err := c.auth.Login(req.Username, req.Password)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, dto.AuthResponse{Token: token})
}