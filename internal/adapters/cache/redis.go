package cache

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"marketflow/pkg/logger"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedis() *RedisCache {
	logger.Info("Starting cache connection...")

	host := os.Getenv("CACHE_HOST")
	port := os.Getenv("CACHE_PORT")
	pass := os.Getenv("CACHE_PASSWORD")

	if host == "" || port == "" {
		logger.Error("CACHE_HOST or CACHE_PORT not set", "host", host, "port", port)
		log.Fatal("missing Redis config")
	}

	addr := fmt.Sprintf("%s:%s", host, port)

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass,
		DB:       0,
	})
	cache := &RedisCache{
		client: client,
	}

	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		_, err := client.Ping(context.Background()).Result()
		if err == nil {
			break
		}

		logger.Warn("failed to connect to Redis", "attempt", i+1, "error", err)
		time.Sleep(2 * time.Second)
	}

	logger.Info("Cache connection established")
	return cache
}

func (r *RedisCache) Close() error {
	logger.Info("closing redis cache")
	return r.client.Close()
}

// SetWithTTL sets a key-value pair in Redis with a specified Time-To-Live (TTL).
// After the ttl duration, Redis will automatically delete the key.
func (r *RedisCache) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	err := r.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		logger.Error("failed to set key with TTL in Redis", "key", key, "error", err)
	} else {
		logger.Info("successfully set key with TTL", "key", key, "ttl", ttl)
	}
	return err
}
