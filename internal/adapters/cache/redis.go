package cache

import (
	"context"
	"fmt"
	"log"
	"marketflow/pkg/config"
	"marketflow/pkg/logger"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedis() *RedisCache {
	logger.Info("Starting cache connection...")

	redisConfig, err := config.LoadRedisConfig()
	if err != nil {
		logger.Error("Error loading Redis config", "error", err)
		log.Fatal(err)
	}

	addr := fmt.Sprintf("%s:%s", redisConfig.Host, redisConfig.Port)

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: redisConfig.Password,
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

func (r *RedisCache) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	err := r.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		logger.Error("failed to set key with TTL in Redis", "key", key, "error", err)
	}
	return err
}
