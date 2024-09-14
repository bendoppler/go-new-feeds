package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
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

var config *Config

// LoadConfig loads configuration from .env file
func LoadConfig() *Config {
	if config == nil {
		err := godotenv.Load()
		if err != nil {
			log.Fatalf("Error loading .env file")
		}

		config = &Config{
			AppName:       getEnv("APP_NAME", "NewsFeedApp"),
			AppPort:       getEnv("APP_PORT", "8080"),
			DBHost:        getEnv("DB_HOST", "localhost"),
			DBPort:        getEnv("DB_PORT", "3306"),
			DBUser:        getEnv("DB_USER", "root"),
			DBPassword:    getEnv("DB_PASSWORD", ""),
			DBName:        getEnv("DB_NAME", "newsfeed"),
			RedisHost:     getEnv("REDIS_HOST", "localhost"),
			RedisPort:     getEnv("REDIS_PORT", "6379"),
			RedisPassword: getEnv("REDIS_PASSWORD", ""),
			JWTSecret:     getEnv("JWT_SECRET", "mysecret"),
		}
	}

	return config
}

// getEnv is a helper function to read an environment variable or return a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
