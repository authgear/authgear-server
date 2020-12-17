package redis

import (
	"context"

	"github.com/go-redis/redis/v8"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type Handle struct {
	pool        *Pool
	cfg         *config.RedisConfig
	credentials *config.RedisCredentials
	logger      *log.Logger
}

func NewHandle(pool *Pool, cfg *config.RedisConfig, credentials *config.RedisCredentials, lf *log.Factory) *Handle {
	return &Handle{
		pool:        pool,
		cfg:         cfg,
		logger:      lf.New("redis-handle"),
		credentials: credentials,
	}
}

func (h *Handle) WithConn(f func(conn *redis.Conn) error) error {
	h.logger.WithFields(map[string]interface{}{
		"max_open_connection":             *h.cfg.MaxOpenConnection,
		"max_idle_connection":             *h.cfg.MaxIdleConnection,
		"idle_connection_timeout_seconds": *h.cfg.IdleConnectionTimeout,
		"max_connection_lifetime_seconds": *h.cfg.MaxConnectionLifetime,
	}).Debug("open redis connection")

	ctx := context.Background()
	conn := h.Client().Conn(ctx)
	defer func() {
		err := conn.Close()
		if err != nil {
			h.logger.WithError(err).Error("failed to close connection")
		}
	}()

	return f(conn)
}

func (h *Handle) Client() *redis.Client {
	return h.pool.Client(h.cfg, h.credentials)
}
