package service

import (
	"context"
	"encoding/json"
	"errors"
	"exchangeapp/internal/apperrors"
	"exchangeapp/internal/client/exchange"
	"exchangeapp/internal/models"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const latestRateCacheTTL = 10 * time.Minute

type rateStore interface {
	Create(exchangeRate *models.ExchangeRate) error
	Upsert(exchangeRate *models.ExchangeRate) error
	Latest() ([]models.ExchangeRate, error)
	FindHistory(fromCurrency string, toCurrency string, startDate time.Time, endDate time.Time) ([]models.ExchangeRate, error)
}

type RateService struct {
	rates          rateStore
	redis          *redis.Client
	exchangeClient exchange.Client
}

type ConvertResult struct {
	Base            string          `json:"base"`
	Quote           string          `json:"quote"`
	Amount          decimal.Decimal `json:"amount"`
	Rate            decimal.Decimal `json:"rate"`
	ConvertedAmount decimal.Decimal `json:"convertedAmount"`
	Date            string          `json:"date"`
}

func NewRateService(rates rateStore, redisClient *redis.Client, exchangeClient exchange.Client) *RateService {
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
		return s.fetchLatestAndCacheAndPersist(ctx, base, quote, cacheKey)
	}
	if err != redis.Nil {
		log.Printf("failed to get latest rate cache %s: %v", cacheKey, err)
		return s.fetchLatestAndCacheAndPersist(ctx, base, quote, cacheKey)
	}

	return s.fetchLatestAndCacheAndPersist(ctx, base, quote, cacheKey)
}

func (s *RateService) Convert(ctx context.Context, base string, quote string, amount decimal.Decimal) (*ConvertResult, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
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
		ConvertedAmount: amount.Mul(latest.Rate),
		Date:            latest.Date,
	}, nil
}

func (s *RateService) History(fromCurrency, toCurrency, startDateStr, endDateStr string) ([]models.ExchangeRate, error) {
	fromCurrency, toCurrency, err := normalizeCurrencyPair(fromCurrency, toCurrency)
	if err != nil {
		return nil, err
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return nil, apperrors.ErrInvalidParams
	}
	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return nil, apperrors.ErrInvalidParams
	}

	if startDate.After(endDate) {
		return nil, apperrors.ErrInvalidParams
	}

	// 限制查询范围不能超过1年，避免过大数据量导致性能问题
	if endDate.Sub(startDate) > 365*24*time.Hour {
		return nil, apperrors.ErrInvalidParams
	}

	return s.rates.FindHistory(fromCurrency, toCurrency, startDate, endDate)
}

func (s *RateService) fetchLatestAndCacheAndPersist(ctx context.Context, base string, quote string, cacheKey string) (*exchange.LatestRate, error) {
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

	if s.rates != nil {
		rateDate, err := time.Parse("2006-01-02", latest.Date)
		if err != nil {
			log.Printf("failed to parse rate date %s, skip persistence: %v", latest.Date, err)
			return latest, nil
		}

		rate := &models.ExchangeRate{
			FromCurrency: latest.Base,
			ToCurrency:   latest.Quote,
			Rate:         latest.Rate,
			RateDate:     rateDate,
		}
		if err := s.rates.Upsert(rate); err != nil {
			log.Printf("failed to upsert exchange rate %s/%s: %v", base, quote, err)
		}
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
