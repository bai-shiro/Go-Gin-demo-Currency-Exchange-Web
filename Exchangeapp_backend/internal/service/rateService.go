package service

import (
	"exchangeapp/internal/models"
	"exchangeapp/internal/repository"

	"github.com/go-redis/redis"
)

type RateService struct {
	rates *repository.RateRepository
	redis *redis.Client
}

func NewRateService(rates *repository.RateRepository, redisClient *redis.Client) *RateService {
	return  &RateService{rates: rates, redis: redisClient}
}

func (s *RateService) Create(exchangeRate *models.ExchangeRate) error {
	return s.rates.Create(exchangeRate)
}

func (s *RateService) Latest() ([]models.ExchangeRate, error) {
	return s.rates.Latest()
}