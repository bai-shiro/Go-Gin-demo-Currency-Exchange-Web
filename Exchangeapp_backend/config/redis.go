package config

import (
	"exchangeapp/global"
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis"
)

func initRedis() {
	addr := "localhost:6379"
	if redisHost := os.Getenv("REDIS_HOST"); redisHost != "" {
		redisPort := os.Getenv("REDIS_PORT")
		addr = fmt.Sprintf("%s:%s", redisHost, redisPort)
	}
	RedisClient := redis.NewClient(&redis.Options{
		Addr: addr,
		DB: 0,
		Password: "",
	})

	_, err := RedisClient.Ping().Result()

	if err != nil {
		log.Fatalf("Failed to connect Redis, err: %v", err)
	}

	global.RedisDB = RedisClient
}