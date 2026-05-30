package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type ExchangeRate struct {
	ID           uint            `gorm:"primaryKey" json:"id"`
	FromCurrency string          `gorm:"size:3;not null;uniqueIndex:idx_rate_pair_date,priority:1" json:"fromCurrency"`
	ToCurrency   string          `gorm:"size:3;not null;uniqueIndex:idx_rate_pair_date,priority:2" json:"toCurrency"`
	Rate         decimal.Decimal `gorm:"type:decimal(20,10);not null" json:"rate"`
	RateDate     time.Time       `gorm:"type:date;not null;uniqueIndex:idx_rate_pair_date,priority:3" json:"rateDate"`
	FetchedAt    time.Time       `gorm:"type:datetime(3);not null" json:"fetchedAt"`
}
