package service

import (
	"context"
	"encoding/json"
	"errors"
	"exchangeapp/internal/apperrors"
	"exchangeapp/internal/client/exchange"
	"exchangeapp/internal/models"
	"exchangeapp/internal/repository"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const latestRateCacheTTL = 10 * time.Minute

type RateService struct {
	rates          *repository.RateRepository
	redis          *redis.Client
	exchangeClient exchange.Client
}

type ConvertResult struct {
	Base            string  `json:"base"`
	Quote           string  `json:"quote"`
	Amount          float64 `json:"amount"`
	Rate            float64 `json:"rate"`
	ConvertedAmount float64 `json:"convertedAmount"`
	Date            string  `json:"date"`
}

func NewRateService(rates *repository.RateRepository, redisClient *redis.Client, exchangeClient exchange.Client) *RateService {
	return &RateService{rates: rates, redis: redisClient, exchangeClient: exchangeClient}
}

func (s *RateService) Create(fromCurrency string, toCurrency string, rateStr string, rateDateStr string) (*models.ExchangeRate, error) {
	fromCurrency, toCurrency, err := normalizeCurrencyPair(fromCurrency, toCurrency)
	if err != nil {
		return nil, err
	}

	rate, err := decimal.NewFromString(rateStr)
	if err != nil || rate.LessThanOrEqual(decimal.Zero) {
		return nil, apperrors.ErrInvalidParams
	}

	rateDate := time.Now()
	if strings.TrimSpace(rateDateStr) != "" {
		rateDate, err = time.Parse("2006-01-02", rateDateStr)
		if err != nil {
			return nil, apperrors.ErrInvalidParams
		}
	}

	exchangeRate := &models.ExchangeRate{FromCurrency: fromCurrency, ToCurrency: toCurrency, Rate: rate, RateDate: rateDate}
	if err := s.rates.Create(exchangeRate); err != nil {
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

func (s *RateService) LatestPair(ctx context.Context, base string, quote string) (*exchange.LatestRate, error) {
	base, quote, err := normalizeCurrencyPair(base, quote)
	if err != nil {
		return nil, err
	}

	cacheKey := latestRateCacheKey(base, quote)
	cacheData, err := s.redis.Get(cacheKey).Result()
	if err == nil {
		var rate exchange.LatestRate
		if err := json.Unmarshal([]byte(cacheData), &rate); err == nil {
			return &rate, nil
		}
		_ = s.redis.Del(cacheKey).Err()
		return s.fetchLatestAndCache(ctx, base, quote, cacheKey)
	}
	if err != redis.Nil {
		log.Printf("failed to get latest rate cache %s: %v", cacheKey, err)
		return s.fetchLatestAndCache(ctx, base, quote, cacheKey)
	}

	return s.fetchLatestAndCache(ctx, base, quote, cacheKey)
}

func (s *RateService) Convert(ctx context.Context, base string, quote string, amount float64) (*ConvertResult, error) {
	if amount <= 0 {
		return nil, apperrors.ErrInvalidParams
	}

	latest, err := s.LatestPair(ctx, base, quote)
	if err != nil {
		return nil, err
	}

	return &ConvertResult{
		Base:            latest.Base,
		Quote:           latest.Quote,
		Amount:          amount,
		Rate:            latest.Rate,
		ConvertedAmount: amount * latest.Rate,
		Date:            latest.Date,
	}, nil
}

func (s *RateService) fetchLatestAndCache(ctx context.Context, base string, quote string, cacheKey string) (*exchange.LatestRate, error) {
	if s.exchangeClient == nil {
		return nil, fmt.Errorf("exchange rate client is nil")
	}

	latest, err := s.exchangeClient.FetchLatest(ctx, base, quote)
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(latest)
	if err == nil {
		_ = s.redis.Set(cacheKey, payload, latestRateCacheTTL).Err()
	}

	return latest, nil
}

func normalizeCurrencyPair(base string, quote string) (string, string, error) {
	base = strings.ToUpper(strings.TrimSpace(base))
	quote = strings.ToUpper(strings.TrimSpace(quote))
	if len(base) != 3 || len(quote) != 3 || base == quote {
		return "", "", apperrors.ErrInvalidParams
	}
	return base, quote, nil
}

func latestRateCacheKey(base string, quote string) string {
	return fmt.Sprintf("rates:latest:%s:%s", base, quote)
}
