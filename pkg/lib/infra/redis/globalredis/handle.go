package globalredis

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
)

type Handle struct {
	*redis.Handle
}

func NewHandle(pool *redis.Pool, cfg *config.RedisEnvironmentConfig, credentials *config.GlobalRedisCredentialsEnvironmentConfig) *Handle {
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
	}
}
