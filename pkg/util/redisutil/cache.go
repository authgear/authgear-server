package redisutil

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
)

// SimpleCmdable is a simplified version of redis.Cmdable.
type SimpleCmdable interface {
	SetEX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
}

type Item struct {
	Key        string
	Expiration time.Duration
	Do         func() ([]byte, error)
}

// Cache is a naive cache that does not prevent multiple clients from
// filling the cache at the same time.
type Cache struct{}

func (c *Cache) Get(ctx context.Context, cmdable SimpleCmdable, item Item) ([]byte, error) {
	bytes, err := cmdable.Get(ctx, item.Key).Bytes()
	if err == nil {
		return bytes, nil
	}

	if !errors.Is(err, redis.Nil) {
		return nil, err
	}

	bytes, err = item.Do()
	if err != nil {
		return nil, err
	}

	_, err = cmdable.SetEX(ctx, item.Key, bytes, item.Expiration).Result()
	if err != nil {
		return nil, err
	}

	return bytes, nil
}
