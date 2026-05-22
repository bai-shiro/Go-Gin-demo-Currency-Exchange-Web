package repository

import (
	"gorm.io/gorm"
)

type Repositories struct {
	Users    *UserRepository
	Articles *ArticleRepository
	Rates    *RateRepository
}

func NewRepositories(db *gorm.DB) *Repositories{
	return &Repositories{
		Users: NewUserRepository(db),
		Articles: NewArticleRepository(db),
		Rates: NewRateRepository(db),
	}
}