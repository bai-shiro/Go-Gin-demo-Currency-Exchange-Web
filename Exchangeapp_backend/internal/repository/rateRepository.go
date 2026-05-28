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
	exchangeRate.Data = time.Now()
	return r.db.Create(exchangeRate).Error
}

func (r *RateRepository) Latest() ([]models.ExchangeRate, error) {
	var exchangeRates []models.ExchangeRate
	if err := r.db.Find(&exchangeRates).Error; err != nil {
		return nil, err
	}
	return exchangeRates, nil
}
