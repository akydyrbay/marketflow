package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"marketflow/internal/domain"
)

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
