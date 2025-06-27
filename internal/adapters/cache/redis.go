package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"marketflow/internal/domain"
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

func (r *RedisCache) GetLatest(ctx context.Context, exchange, pair string) (domain.PriceUpdate, error) {
	key := fmt.Sprintf("latest:%s:%s", exchange, pair)
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		logger.Warn("no data in redis", "key", key)
		return domain.PriceUpdate{}, fmt.Errorf("no data for %s", key)
	}
	if err != nil {
		logger.Warn("redis get error, using fallback", "key", key, "error", err)
		return domain.PriceUpdate{}, fmt.Errorf("redis unavailable: %w", err)
	}
	var update domain.PriceUpdate
	if err := json.Unmarshal([]byte(val), &update); err != nil {
		logger.Error("unmarshal error", "key", key, "error", err)
		return domain.PriceUpdate{}, fmt.Errorf("unmarshal error: %w", err)
	}
	logger.Info("got latest price", "key", key, "price", update.Price)
	return update, nil
}
