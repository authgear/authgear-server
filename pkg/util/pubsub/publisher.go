package pubsub

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type RedisPool interface {
	Get() *redis.Client
}

type Publisher struct {
	RedisPool RedisPool
}

func (p *Publisher) Publish(ctx context.Context, channelName string, data []byte) error {
	redisClient := p.RedisPool.Get()
	cmd := redisClient.Publish(ctx, channelName, string(data))
	return cmd.Err()
}
