package analyticredis

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type Handle struct {
	*redis.Handle
}

func NewHandle(pool *redis.Pool, cfg *config.RedisConfig, credentials *config.AnalyticRedisCredentials, lf *log.Factory) *Handle {
	if credentials == nil {
		return nil
	}

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
			lf.New("analyticredis-handle"),
		),
	}
}
