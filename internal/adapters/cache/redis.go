package cache

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
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

func (r *RedisCache) StoreLatestPrice(ctx context.Context, update domain.PriceUpdate) error {
	key := fmt.Sprintf("latest:%s:%s", update.Exchange, update.Pair)
	value := fmt.Sprintf("%f", update.Price)

	// Store with 5 minute expiration
	return r.client.Set(ctx, key, value, 5*time.Minute).Err()
}

func (r *RedisCache) GetLatest(ctx context.Context, exchange, pair string) (domain.PriceUpdate, error) {
	key := fmt.Sprintf("latest:%s:%s", exchange, pair)
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return domain.PriceUpdate{}, err
	}

	price, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return domain.PriceUpdate{}, err
	}

	return domain.PriceUpdate{
		Exchange: exchange,
		Pair:     pair,
		Price:    price,
		Time:     time.Now(),
	}, nil
}
