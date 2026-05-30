package repository

import (
	"exchangeapp/internal/models"
	"time"

	"gorm.io/gorm"
)

type RateRepository struct {
	db *gorm.DB
}

func NewRateRepository(db *gorm.DB) *RateRepository {
	return &RateRepository{db: db}
}

func (r *RateRepository) Create(exchangeRate *models.ExchangeRate) error {
	now := time.Now()
	if exchangeRate.FetchedAt.IsZero() {
		exchangeRate.FetchedAt = now
	}
	if exchangeRate.RateDate.IsZero() {
		y, m, d := now.Date()
		exchangeRate.RateDate = time.Date(y, m, d, 0, 0, 0, 0, now.Location())
	}
	return r.db.Create(exchangeRate).Error
}

func (r *RateRepository) Latest() ([]models.ExchangeRate, error) {
	var exchangeRates []models.ExchangeRate
	if err := r.db.Order("rate_date DESC, fetched_at DESC").Find(&exchangeRates).Error; err != nil {
		return nil, err
	}
	return exchangeRates, nil
}
