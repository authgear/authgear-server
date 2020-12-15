package pubsub

import (
	"context"
)

type Publisher struct {
	RedisPool RedisPool
}

func (p *Publisher) Publish(channelName string, data []byte) error {
	ctx := context.Background()
	redisClient := p.RedisPool.Get()
	cmd := redisClient.Publish(ctx, channelName, string(data))
	return cmd.Err()
}
