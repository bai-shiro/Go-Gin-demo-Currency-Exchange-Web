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

	tokens, err := c.auth.Register(req.Username, req.Password)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, toAuthResponse(tokens))
}

func (c *AuthController) Login(ctx *gin.Context) {
	var req dto.LoginRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, apperrors.ErrInvalidParams)
		return
	}

	tokens, err := c.auth.Login(req.Username, req.Password)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, toAuthResponse(tokens))
}

func (c *AuthController) Refresh(ctx *gin.Context) {
	var req dto.RefreshTokenRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, apperrors.ErrInvalidParams)
		return
	}

	tokens, err := c.auth.Refresh(req.RefreshToken)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, toAuthResponse(tokens))
}

func (c *AuthController) Logout(ctx *gin.Context) {
	var req dto.LogoutRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, apperrors.ErrInvalidParams)
		return
	}

	if err := c.auth.Logout(req.RefreshToken); err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, gin.H{"loggedOut": true})
}

func toAuthResponse(tokens *service.AuthTokens) dto.AuthResponse {
	return dto.AuthResponse{
		AccessToken:           tokens.AccessToken,
		RefreshToken:          tokens.RefreshToken,
		AccessTokenExpiresIn:  tokens.AccessTokenExpiresIn,
		RefreshTokenExpiresIn: tokens.RefreshTokenExpiresIn,
	}
}
