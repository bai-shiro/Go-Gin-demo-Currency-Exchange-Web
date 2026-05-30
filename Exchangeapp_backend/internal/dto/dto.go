package dto

import "github.com/shopspring/decimal"

type ArticleRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
	Preview string `json:"preview" binding:"required"`
}

type CreateRateRequest struct {
	FromCurrency string `json:"fromCurrency" binding:"required,len=3"`
	ToCurrency   string `json:"toCurrency" binding:"required,len=3"`
	Rate         string `json:"rate" binding:"required"`
	RateDate     string `json:"rateDate" binding:"omitempty"` // 格式：2006-01-02
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type ArticleResponse struct {
	ID      uint   `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Preview string `json:"preview"`
}

type RateResponse struct {
	FromCurrency string `json:"fromCurrency"`
	ToCurrency   string `json:"toCurrency"`
	Rate         string `json:"rate"`
	RateDate     string `json:"rateDate"`
	FetchedAt    string `json:"fetchedAt"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

type LatestRateResponse struct {
	Base  string          `json:"base"`
	Quote string          `json:"quote"`
	Rate  decimal.Decimal `json:"rate"`
	Date  string          `json:"date"`
}

type ConvertResponse struct {
	Base            string          `json:"base"`
	Quote           string          `json:"quote"`
	Amount          decimal.Decimal `json:"amount"`
	Rate            decimal.Decimal `json:"rate"`
	ConvertedAmount decimal.Decimal `json:"convertedAmount"`
	Date            string          `json:"date"`
}
