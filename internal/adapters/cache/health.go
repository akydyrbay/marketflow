package cache

import (
	"context"
)

func (c *RedisCache) CheckHealth() error {
	_, err := c.client.Ping(context.Background()).Result()
	if err != nil {
		return err
	}
	return nil
}
