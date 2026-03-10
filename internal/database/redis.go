package database

import (
	"context"
	"fmt"
	"go-fiber-stater-kit/config"
	"log"

	"github.com/gofiber/storage/redis/v3"
	goredis "github.com/redis/go-redis/v9"
)

// NewRedis creates a Fiber-compatible Redis storage (for rate limiting, etc.)
func NewRedis(cfg *config.Config) *redis.Storage {
	store := redis.New(redis.Config{
		Host:     cfg.RedisHost,
		Port:     cfg.RedisPort,
		Password: cfg.RedisPassword,
		Database: 0,
		Reset:    false,
	})

	log.Println("Redis storage connected")
	return store
}

// NewRedisClient creates a go-redis client for token blacklisting and caching
func NewRedisClient(cfg *config.Config) *goredis.Client {
	client := goredis.NewClient(&goredis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       0,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatalf("Cannot connect to Redis: %v", err)
	}

	log.Println("Redis client connected")
	return client
}
