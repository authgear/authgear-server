package appredis

import (
	goredis "github.com/redis/go-redis/v9"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
)

type Handle struct {
	*redis.Handle

	hub *redis.Hub
}

func NewHandle(pool *redis.Pool, hub *redis.Hub, cfg *config.RedisEnvironmentConfig, credentials *config.RedisCredentials) *Handle {
	return &Handle{
		Handle: redis.NewHandle(
			pool,
			redis.ConnectionOptions{
				RedisURL:              credentials.RedisURL,
				MaxOpenConnection:     &cfg.MaxOpenConnection,
				MaxIdleConnection:     &cfg.MaxIdleConnection,
				IdleConnectionTimeout: &cfg.IdleConnectionTimeout,
				MaxConnectionLifetime: &cfg.MaxConnectionLifetime,
			},
		),
		hub: hub,
	}
}

func (h *Handle) Subscribe(channelName string) (chan *goredis.Message, func()) {
	sub := h.hub.Subscribe(&h.ConnectionOptions, channelName)
	return sub.MessageChannel, sub.Cancel
}
