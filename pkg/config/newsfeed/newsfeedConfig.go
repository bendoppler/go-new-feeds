package newsfeed

import (
	"github.com/spf13/viper"
	"log"
)

type NewsfeedConfig struct {
	AppName       string
	AppPort       string
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	RedisHost     string
	RedisPort     string
	RedisPassword string
	JWTSecret     string
}

var config *NewsfeedConfig

func LoadNewsfeedConfig() *NewsfeedConfig {
	viper.SetConfigType("env")
	viper.SetConfigFile(".env.newsfeed")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error loading .env.newsfeed: %v", err)
	}
	if config == nil {
		config = &NewsfeedConfig{
			AppName:       getEnv("APP_NAME", "NewsFeedService"),
			AppPort:       getEnv("APP_PORT", "8081"),
			DBHost:        getEnv("DB_HOST", "localhost"),
			DBPort:        getEnv("DB_PORT", "3306"),
			DBUser:        getEnv("DB_USER", "root"),
			DBPassword:    getEnv("DB_PASSWORD", ""),
			DBName:        getEnv("DB_NAME", "newsfeed"),
			RedisHost:     getEnv("REDIS_HOST", "localhost"),
			RedisPort:     getEnv("REDIS_PORT", "6379"),
			RedisPassword: getEnv("REDIS_PASSWORD", ""),
			JWTSecret:     getEnv("JWT_SECRET", ""),
		}
	}

	return config
}

func getEnv(key, defaultValue string) string {
	if value := viper.GetString(key); value != "" {
		return value
	}
	return defaultValue
}
