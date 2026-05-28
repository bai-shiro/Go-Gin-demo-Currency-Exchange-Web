package service

import (
	"exchangeapp/internal/client/exchange"
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
	exchangeClient := exchange.NewFrankfurterClient(exchange.DefaultFrankfurterBaseURL, 3*time.Second)
	return &Services{
		Articles: NewArticleService(repos.Articles, redisClient),
		Rates:    NewRateService(repos.Rates, redisClient, exchangeClient),
		Auth:     NewAuthService(repos.Users, jwtSecret, jwtTTL),
	}
}
