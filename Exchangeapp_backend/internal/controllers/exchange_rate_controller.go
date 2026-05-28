package controllers

import (
	"exchangeapp/internal/apperrors"
	"exchangeapp/internal/dto"
	"exchangeapp/internal/models"
	"exchangeapp/internal/response"
	"exchangeapp/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RateController struct {
	rates *service.RateService
}

func NewRateController(rates *service.RateService) *RateController {
	return &RateController{rates: rates}
}

func (c *RateController) Create(ctx *gin.Context) {
	var req dto.CreateRateRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, apperrors.ErrInvalidParams)
		return
	}

	exchangeRate, err := c.rates.Create(req.FromCurrency, req.ToCurrency, req.Rate)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Created(ctx, toRateResponse(exchangeRate))
}

func (c *RateController) Latest(ctx *gin.Context) {
	exchangeRates, err := c.rates.Latest()
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, toRateResponses(exchangeRates))
}

func (c *RateController) LatestPair(ctx *gin.Context) {
	base := ctx.Query("base")
	quote := ctx.Query("quote")

	latest, err := c.rates.LatestPair(ctx.Request.Context(), base, quote)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, dto.LatestRateResponse{
		Base:  latest.Base,
		Quote: latest.Quote,
		Rate:  latest.Rate,
		Date:  latest.Date,
	})
}

func (c *RateController) Convert(ctx *gin.Context) {
	base := ctx.Query("base")
	quote := ctx.Query("quote")
	amount, err := strconv.ParseFloat(ctx.Query("amount"), 64)
	if err != nil {
		response.Error(ctx, apperrors.ErrInvalidParams)
		return
	}

	converted, err := c.rates.Convert(ctx.Request.Context(), base, quote, amount)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, dto.ConvertResponse{
		Base:            converted.Base,
		Quote:           converted.Quote,
		Amount:          converted.Amount,
		Rate:            converted.Rate,
		ConvertedAmount: converted.ConvertedAmount,
		Date:            converted.Date,
	})
}

func toRateResponse(rate *models.ExchangeRate) dto.RateResponse {
	return dto.RateResponse{
		FromCurrency: rate.FromCurrency,
		ToCurrency:   rate.ToCurrency,
		Rate:         rate.Rate,
	}
}

func toRateResponses(rates []models.ExchangeRate) []dto.RateResponse {
	res := make([]dto.RateResponse, len(rates))
	for i := range rates {
		res[i] = toRateResponse(&rates[i])
	}
	return res
}
