package service

import (
	"exchangeapp/internal/repository"
	"time"

	"github.com/go-redis/redis"
)

type Services struct {
	Articles *ArticleService
	Rates    *RateService
	Auth     *AuthService
}

func NewServices(repos *repository.Repositories, redisClient *redis.Client, jwtSecret string, jwtTTL time.Duration) *Services {
	return &Services{
		Articles: NewArticleService(repos.Articles, redisClient),
		Rates:    NewRateService(repos.Rates, redisClient),
		Auth:     NewAuthService(repos.Users, jwtSecret, jwtTTL),
	}
}
