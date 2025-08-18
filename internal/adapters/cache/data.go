package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"marketflow/internal/domain"
	"time"
)

// SaveLatestData saves the most recent data points to the cache with a 5-minute expiration.
func (c *RedisCache) SaveLatestData(latestData map[string]domain.Data) error {
	expiration := 5 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	for key, value := range latestData {
		jsonData, err := json.Marshal(value)
		if err != nil {
			return err
		}

		err = c.SetWithTTL(ctx, key, jsonData, expiration)
		if err != nil {
			return err
		}
	}

	return nil
}

// SaveAggregatedData saves aggregated data to the cache with a 5-minute expiration.
func (c *RedisCache) SaveAggregatedData(aggregatedData map[string]domain.ExchangeData) error {
	expiration := 5 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	for key, value := range aggregatedData {
		jsonData, err := json.Marshal(value)
		if err != nil {
			return err
		}

		err = c.SetWithTTL(ctx, key, jsonData, expiration)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *RedisCache) LatestData(exchange, symbol string) (domain.Data, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	key := "latest " + exchange + " " + symbol
	res, err := c.client.Get(ctx, key).Result()

	fmt.Println("DEBUG: ", res)

	if err != nil {
		return domain.Data{}, err
	}
	raw := domain.Data{}
	if err := json.Unmarshal([]byte(res), &raw); err != nil {
		return domain.Data{}, err
	}

	return raw, nil
}
