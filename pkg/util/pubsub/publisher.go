package pubsub

import (
	"context"

	"github.com/go-redis/redis/v8"
)

type RedisPool interface {
	Get() *redis.Client
}

type Publisher struct {
	RedisPool RedisPool
}

func (p *Publisher) Publish(channelName string, data []byte) error {
	ctx := context.Background()
	redisClient := p.RedisPool.Get()
	cmd := redisClient.Publish(ctx, channelName, string(data))
	return cmd.Err()
}
