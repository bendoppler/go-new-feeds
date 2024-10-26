package cache

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"news-feed/pkg/config/webApp"
	"time"
)

var redisClient *redis.Client

func init() {
	// Initialize Redis client
	redisClient = newRedisClient()
}

func newRedisClient() *redis.Client {
	cfg := webApp.LoadConfig()
	redisHost := cfg.RedisHost
	redisPort := cfg.RedisPort
	addr := fmt.Sprintf("%s:%s", redisHost, redisPort)

	client := redis.NewClient(
		&redis.Options{
			Addr:         addr,
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
