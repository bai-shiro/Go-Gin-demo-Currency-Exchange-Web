package service

import (
	"errors"
	"exchangeapp/internal/apperrors"
	"exchangeapp/internal/models"
	"exchangeapp/internal/repository"

	"github.com/go-redis/redis"
	"gorm.io/gorm"
)

type RateService struct {
	rates *repository.RateRepository
	redis *redis.Client
}

func NewRateService(rates *repository.RateRepository, redisClient *redis.Client) *RateService {
	return &RateService{rates: rates, redis: redisClient}
}

func (s *RateService) Create(fromCurrency string, toCurrency string, rate float64) (*models.ExchangeRate, error) {
	exchangeRate := &models.ExchangeRate{FromCurrency: fromCurrency, ToCurrency: toCurrency, Rate: rate}
	err := s.rates.Create(exchangeRate)
	if err != nil {
		return nil, err
	}
	return exchangeRate, nil
}

func (s *RateService) Latest() ([]models.ExchangeRate, error) {
	rates, err := s.rates.Latest()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrNotFound
	} else if err != nil {
		return nil, err
	}
	return rates, nil
}
