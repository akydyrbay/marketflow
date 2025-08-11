package cache

import (
	"context"
	"encoding/json"
	"marketflow/internal/domain"
	"time"
)

func (c *RedisCache) SaveLatestData(latestData map[string]domain.Data) error {
	for key, value := range latestData {
		jsonData, err := json.Marshal(value)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

		err = c.client.Set(ctx, key, jsonData, 0).Err()
		cancel()
		if err != nil {
			return err
		}
	}

	return nil
}
