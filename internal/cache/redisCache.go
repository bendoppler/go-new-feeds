package cache

import (
	"context"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"log"
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

	client := redis.NewClient(
		&redis.Options{
			Addr:         "localhost:6379",
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
