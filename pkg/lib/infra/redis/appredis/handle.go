package appredis

import (
	goredis "github.com/go-redis/redis/v8"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type Handle struct {
	*redis.Handle

	hub *redis.Hub
}

func NewHandle(pool *redis.Pool, hub *redis.Hub, cfg *config.RedisConfig, credentials *config.RedisCredentials, lf *log.Factory) *Handle {
	return &Handle{
		Handle: redis.NewHandle(
			pool,
			redis.ConnectionOptions{
				RedisURL:              credentials.RedisURL,
				MaxOpenConnection:     cfg.MaxOpenConnection,
				MaxIdleConnection:     cfg.MaxIdleConnection,
				IdleConnectionTimeout: cfg.IdleConnectionTimeout,
				MaxConnectionLifetime: cfg.MaxConnectionLifetime,
			},
			lf.New("appredis-handle"),
		),
		hub: hub,
	}
}

func (h *Handle) Subscribe(channelName string) (chan *goredis.Message, func()) {
	sub := h.hub.Subscribe(&h.ConnectionOptions, channelName)
	return sub.MessageChannel, sub.Cancel
}
