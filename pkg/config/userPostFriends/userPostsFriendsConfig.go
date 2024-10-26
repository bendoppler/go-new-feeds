package userPostFriends

import (
	"github.com/spf13/viper"
	"log"
)

type UserPostFriendsConfig struct {
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

var config *UserPostFriendsConfig

func LoadUserPostFriendsConfig() *UserPostFriendsConfig {
	viper.SetConfigType("env")
	viper.SetConfigFile(".env.userPostFriends")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error loading .env.userPostFriends: %v", err)
	}
	if config == nil {
		config = &UserPostFriendsConfig{
			AppName:       getEnv("APP_NAME", "UserPostFriends"),
			AppPort:       getEnv("APP_PORT", "8082"),
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
