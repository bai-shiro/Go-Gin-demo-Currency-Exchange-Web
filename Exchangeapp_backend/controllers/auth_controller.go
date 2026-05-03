package controllers

import (
	"exchangeapp/global"
	"exchangeapp/models"
	"exchangeapp/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Register(ctx *gin.Context) {
	var user models.User

	if err := ctx.ShouldBindJSON(&user); err != err {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"Error" : err.Error(),
		})
		return
	}

	hashedPwd, err := utils.HashPassword(user.Password)
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"Error" : err.Error(),
		})
		return
	}

	user.Password = hashedPwd

	token, err := utils.GenerateJWT(user.Username)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"Error" : err.Error(),
		})
		return
	}

	if err := global.Db.AutoMigrate(&user); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"Error" : err.Error()})
		return
	}

	if err := global.Db.Create(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"Error" : err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"token" : token})
	
}