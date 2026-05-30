package repository

import (
	"exchangeapp/internal/models"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RateRepository struct {
	db *gorm.DB
}

func NewRateRepository(db *gorm.DB) *RateRepository {
	return &RateRepository{db: db}
}

func (r *RateRepository) Create(exchangeRate *models.ExchangeRate) error {
	prepareExchangeRateForWrite(exchangeRate)
	return r.db.Create(exchangeRate).Error
}

func (r *RateRepository) Upsert(exchangeRate *models.ExchangeRate) error {
	prepareExchangeRateForWrite(exchangeRate)
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "from_currency"},
			{Name: "to_currency"},
			{Name: "rate_date"},
		},
		DoUpdates: clause.AssignmentColumns([]string{"rate", "fetched_at"}),
	}).Create(exchangeRate).Error
}

func (r *RateRepository) Latest() ([]models.ExchangeRate, error) {
	var exchangeRates []models.ExchangeRate
	if err := r.db.Order("rate_date DESC, fetched_at DESC").Find(&exchangeRates).Error; err != nil {
		return nil, err
	}
	return exchangeRates, nil
}

func (r *RateRepository) FindLatestPair(fromCurrency string, toCurrency string) (*models.ExchangeRate, error) {
	var exchangeRate models.ExchangeRate
	if err := r.db.
		Where("from_currency = ? AND to_currency = ?", fromCurrency, toCurrency).
		Order("rate_date DESC, fetched_at DESC").
		First(&exchangeRate).Error; err != nil {
		return nil, err
	}
	return &exchangeRate, nil
}

func (r *RateRepository) FindHistory(fromCurrency string, toCurrency string, startDate time.Time, endDate time.Time) ([]models.ExchangeRate, error) {
	var exchangeRates []models.ExchangeRate
	startDate = dateOnly(startDate)
	endDate = dateOnly(endDate)

	if err := r.db.
		Where("from_currency = ? AND to_currency = ? AND rate_date BETWEEN ? AND ?", fromCurrency, toCurrency, startDate, endDate).
		Order("rate_date ASC").
		Find(&exchangeRates).Error; err != nil {
		return nil, err
	}
	return exchangeRates, nil
}

func prepareExchangeRateForWrite(exchangeRate *models.ExchangeRate) {
	now := time.Now()
	if exchangeRate.FetchedAt.IsZero() {
		exchangeRate.FetchedAt = now
	}
	if exchangeRate.RateDate.IsZero() {
		exchangeRate.RateDate = dateOnly(now)
		return
	}
	exchangeRate.RateDate = dateOnly(exchangeRate.RateDate)
}

func dateOnly(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}
