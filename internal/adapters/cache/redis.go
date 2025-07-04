package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"

	"marketflow/internal/domain"
	"marketflow/pkg/logger"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedis() *RedisCache {
	slog.Info("Starting cache connection...")

	client := redis.NewClient(&redis.Options{Addr: os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"), Password: os.Getenv("REDIS_PASSWORD"), DB: 0})
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to connect cache memory: %s", err.Error())
	}

	slog.Info("Cache connection finished...")
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
