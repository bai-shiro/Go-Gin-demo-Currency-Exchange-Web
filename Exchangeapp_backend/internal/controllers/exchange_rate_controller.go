package controllers

import (
	"errors"
	"exchangeapp/internal/models"
	"exchangeapp/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RateController struct {
	rates *service.RateService
}

func NewRateController(rates *service.RateService) *RateController {
	return  &RateController{rates: rates}
}

func (c *RateController) Create(ctx *gin.Context) {
	var exchangeRate models.ExchangeRate

	if err := ctx.ShouldBindJSON(&exchangeRate); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"Error" : err.Error()})
		return
	}

	if err := c.rates.Create(&exchangeRate); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"Error" : err.Error()})
	}

	ctx.JSON(http.StatusCreated, exchangeRate)
}

func (c *RateController) Latest(ctx *gin.Context) {
	exchangeRates, err := c.rates.Latest()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"Error" : err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"Error" : err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, exchangeRates)
}