package redis

import (
	redigo "github.com/gomodule/redigo/redis"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type Conn = redigo.Conn

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

func (h *Handle) WithConn(f func(conn Conn) error) error {
	h.logger.WithFields(map[string]interface{}{
		"max_open_connection":             *h.cfg.MaxOpenConnection,
		"max_idle_connection":             *h.cfg.MaxIdleConnection,
		"idle_connection_timeout_seconds": *h.cfg.IdleConnectionTimeout,
		"max_connection_lifetime_seconds": *h.cfg.MaxConnectionLifetime,
	}).Debug("open redis connection")

	conn := h.Pool().Get()
	defer func() {
		err := conn.Close()
		if err != nil {
			h.logger.WithError(err).Error("failed to close connection")
		}
	}()

	return f(conn)
}

func (h *Handle) Pool() *redigo.Pool {
	return h.pool.Open(h.cfg, h.credentials)
}
