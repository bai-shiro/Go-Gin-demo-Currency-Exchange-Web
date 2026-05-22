package controllers

import (
	"errors"
	"exchangeapp/internal/models"
	"exchangeapp/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	auth *service.AuthService
}

func NewAuthController(auth *service.AuthService) *AuthController {
	return &AuthController{auth: auth}
}

func (c *AuthController) Register(ctx *gin.Context) {
	var user models.User

	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"Error" : err.Error(),
		})
		return
	}

	token, err := c.auth.Register(&user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"Error" : err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"token" : token})
	
}

func (c *AuthController) Login(ctx *gin.Context) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"Error" : err.Error()})
		return
	}

	token, err := c.auth.Login(input.Username, input.Password)

	if errors.Is(err, errors.New("incorrect credentials")) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"Error" : "incorrect credentials"})
		return
	} else if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{ "Error" : err.Error() })
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"token" : token})
}