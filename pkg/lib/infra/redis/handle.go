package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type Handle struct {
	pool *Pool
	hub  *Hub

	cfg         *config.RedisConfig
	credentials *config.RedisCredentials
	logger      *log.Logger
}

func NewHandle(pool *Pool, hub *Hub, cfg *config.RedisConfig, credentials *config.RedisCredentials, lf *log.Factory) *Handle {
	return &Handle{
		pool:        pool,
		hub:         hub,
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

func (h *Handle) Subscribe(channelName string) (chan *redis.Message, func()) {
	sub := h.hub.Subscribe(h.cfg, h.credentials, channelName)
	return sub.MessageChannel, sub.Cancel
}

func (h *Handle) Client() *redis.Client {
	return h.pool.Client(h.cfg, h.credentials)
}

func (h *Handle) NewMutex(name string) *redsync.Mutex {
	redsyncInstance := h.pool.instance(h.cfg, h.credentials).Redsync
	mutex := redsyncInstance.NewMutex(
		name,
		redsync.WithExpiry(5*time.Second),
		redsync.WithTries(5),
	)
	return mutex
}
