package webApp

import (
	"github.com/spf13/viper"
	"log"
)

type WebAppConfig struct {
	AppName             string
	AppPort             string
	NewsfeedAppPort     string
	PostUserFriendsPort string
	JWTSecret           string
	DBHost              string
	DBPort              string
	DBUser              string
	DBPassword          string
	DBName              string
	RedisHost           string
	RedisPort           string
	MinIOEndpoint       string
	MinIOAccessKey      string
	MinIOSecretKey      string
	MinIOBucket         string
}

var config *WebAppConfig

// LoadConfig loads configuration from .env file
func LoadConfig() *WebAppConfig {
	viper.SetConfigType("env") // Set the config type
	viper.SetConfigFile(".env.webApp")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error loading .env.webApp: %v", err)
	}
	if config == nil {
		config = &WebAppConfig{
			AppName:             getEnv("APP_NAME", "WebApp"),
			AppPort:             getEnv("APP_PORT", "8080"),
			NewsfeedAppPort:     getEnv("NEWSFEED_APP_PORT", "8081"),
			PostUserFriendsPort: getEnv("POST_USER_FRIENDS_PORT", "8082"),
			JWTSecret:           getEnv("JWTSecret", ""),
			DBHost:              getEnv("DB_HOST", "localhost"),
			DBPort:              getEnv("DB_PORT", "5432"),
			DBUser:              getEnv("DB_USER", ""),
			DBPassword:          getEnv("DB_PASSWORD", ""),
			DBName:              getEnv("DB_NAME", ""),
			RedisHost:           getEnv("REDIS_HOST", "localhost"),
			RedisPort:           getEnv("REDIS_PORT", "6379"),
			MinIOEndpoint:       getEnv("MINIO_ENDPOINT", ""),
			MinIOAccessKey:      getEnv("MINIO_ACCESS_KEY", ""),
			MinIOSecretKey:      getEnv("MINIO_SECRET_KEY", ""),
			MinIOBucket:         getEnv("MINIO_BUCKET", ""),
		}
	}

	return config
}

// getEnv is a helper function to read an environment variable or return a default value
func getEnv(key, defaultValue string) string {
	if value := viper.GetString(key); value != "" {
		return value
	}
	return defaultValue
}
