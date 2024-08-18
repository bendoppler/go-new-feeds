package cache

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"log"
	"os"
)

var redisClient *redis.Client
var ctx = context.Background()

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

	client := redis.NewClient(
		&redis.Options{
			Addr:     "localhost:6379",
			Password: password,
			DB:       0,
		},
	)

	// Ping the Redis server to test connection
	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}

	return client
}

func GetRedisClient() *redis.Client {
	return redisClient
}
