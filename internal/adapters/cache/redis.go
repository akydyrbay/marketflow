package cache

import (
	"context"
	"time"

	"marketflow/pkg/logger"

	"github.com/go-redis/redis/v8"
)

type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedis(db int, addr, password string, ttl time.Duration) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	cache := &RedisCache{
		client: client,
		ttl:    ttl,
	}
	for i := 0; i < 3; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := client.Ping(ctx).Err(); err != nil {
			logger.Error("failed to ping redis", "attempt", i+1, "error", err)
			if i == 2 {
				logger.Warn("redis connection failed after retries, proceeding with fallback")
			}
			time.Sleep(time.Second * time.Duration(i+1))
		} else {
			logger.Info("redis connection established")
			break
		}
	}
	return cache
}

func (r *RedisCache) Close() error {
	logger.Info("closing redis cache")
	return r.client.Close()
}
