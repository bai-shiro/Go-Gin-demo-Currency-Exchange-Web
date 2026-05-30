package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Config struct {
	App struct {
		Name string `mapstructure:"name"`
		Port string `mapstructure:"port"`
	}
	Database struct {
		Dsn          string `mapstructure:"dsn"`
		MaxIdleConns int    `mapstructure:"maxIdleConns"`
		MaxOpenConns int    `mapstructure:"maxOpenConns"`
		AutoMigrate  bool   `mapstructure:"autoMigrate"`
	}
	Cache struct {
		Addr     string `mapstructure:"addr"`
		Password string `mapstructure:"password"`
		DB       int    `mapstructure:"db"`
	} `mapstructure:"cache"`

	JWT struct {
		Secret string        `mapstructure:"secret"`
		TTL    time.Duration `mapstructure:"ttl"`
	} `mapstructure:"jwt"`

	Db      *gorm.DB
	RedisDB *redis.Client
}

func InitConfig() (*Config, error) {
	appConfig, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	initDB(appConfig)
	initRedis(appConfig)

	return appConfig, nil
}

func LoadConfig() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yml")
	v.AddConfigPath("./configs")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("Error in reading yml config file: %v", err)
	}

	appConfig := &Config{}

	if err := v.Unmarshal(appConfig); err != nil {
		return nil, fmt.Errorf("Unable to decode into struct: %v", err)
	}

	return appConfig, nil
}

func initDB(appConfig *Config) {
	dsn := ApplyDatabaseDSNEnv(appConfig.Database.Dsn)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatalf("Failed to init database, err: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to config database, err: %v", err)
	}

	sqlDB.SetMaxIdleConns(appConfig.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(appConfig.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	appConfig.Db = db
	appConfig.Database.Dsn = dsn
}

func initRedis(appConfig *Config) {
	addr := appConfig.Cache.Addr
	if redisHost := os.Getenv("REDIS_HOST"); redisHost != "" {
		redisPort := os.Getenv("REDIS_PORT")
		addr = fmt.Sprintf("%s:%s", redisHost, redisPort)
	}
	RedisClient := redis.NewClient(&redis.Options{
		Addr:     addr,
		DB:       appConfig.Cache.DB,
		Password: appConfig.Cache.Password,
	})

	_, err := RedisClient.Ping().Result()

	if err != nil {
		log.Fatalf("Failed to connect Redis, err: %v", err)
	}

	appConfig.RedisDB = RedisClient
}

func ApplyDatabaseDSNEnv(dsn string) string {
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		dbPort := os.Getenv("DB_PORT")
		dbUser := os.Getenv("DB_USER")
		dbPassword := os.Getenv("DB_PASSWORD")
		dbNAME := os.Getenv("DB_NAME")
		if dbPort == "" {
			dbPort = "3306"
		}
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbUser, dbPassword, dbHost, dbPort, dbNAME)
	}
	return dsn
}
