package cache_redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yohnnn/booking_service/internal/models"
)

type ConcertCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewConcertCache(client *redis.Client, ttl time.Duration) *ConcertCache {
	return &ConcertCache{
		client: client,
		ttl:    ttl,
	}
}

func (c *ConcertCache) Get(ctx context.Context) ([]models.Concert, error) {
	val, err := c.client.Get(ctx, "concerts:all").Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var concerts []models.Concert
	if err := json.Unmarshal([]byte(val), &concerts); err != nil {
		return nil, err
	}

	return concerts, nil
}

func (c *ConcertCache) Set(ctx context.Context, concerts []models.Concert) error {
	data, err := json.Marshal(concerts)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, "concerts:all", data, c.ttl).Err()
}

func (c *ConcertCache) Delete(ctx context.Context) error {
	return c.client.Del(ctx, "concerts:all").Err()
}
