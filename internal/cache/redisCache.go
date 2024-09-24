package cache

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
	"time"
)

var redisClient *redis.Client

func init() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Initialize Redis client
	redisClient = newRedisClient()
}

func newRedisClient() *redis.Client {
	password := os.Getenv("REDIS_PASSWORD")
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	addr := fmt.Sprintf("%s:%s", redisHost, redisPort)

	client := redis.NewClient(
		&redis.Options{
			Addr:         addr,
			Password:     password,
			DB:           0,
			PoolSize:     300,              // Further increase if needed
			MinIdleConns: 50,               // Minimum idle connections
			DialTimeout:  10 * time.Second, // Reduce timeouts
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		},
	)

	// Ping the Redis server to test connection
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}

	return client
}

func GetRedisClient() *redis.Client {
	return redisClient
}
