package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"marketflow/internal/domain"
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
	var client *redis.Client
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		client = redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: pass,
			DB:       0,
		})

		_, err := client.Ping(context.Background()).Result()
		if err == nil {
			break
		}

		logger.Warn("failed to connect to Redis", "attempt", i+1, "error", err)
		time.Sleep(2 * time.Second)
	}

	logger.Info("Cache connection established")
	return &RedisCache{client: client}
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
