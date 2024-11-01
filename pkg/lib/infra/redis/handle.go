package redis

import (
	"context"
	"time"

	"github.com/go-redsync/redsync/v4"
	goredis "github.com/redis/go-redis/v9"

	"github.com/authgear/authgear-server/pkg/util/log"
)

type Handle struct {
	pool *Pool

	ConnectionOptions ConnectionOptions
	logger            *log.Logger
}

func NewHandle(pool *Pool, connectionOptions ConnectionOptions, logger *log.Logger) *Handle {
	return &Handle{
		pool:              pool,
		ConnectionOptions: connectionOptions,
		logger:            logger,
	}
}

func (h *Handle) WithConnContext(ctx context.Context, do func(ctx context.Context, conn Redis_6_0_Cmdable) error) error {
	h.logger.WithFields(map[string]interface{}{
		"max_open_connection":             *h.ConnectionOptions.MaxOpenConnection,
		"max_idle_connection":             *h.ConnectionOptions.MaxIdleConnection,
		"idle_connection_timeout_seconds": *h.ConnectionOptions.IdleConnectionTimeout,
		"max_connection_lifetime_seconds": *h.ConnectionOptions.MaxConnectionLifetime,
	}).Debug("open redis connection")

	conn := h.Client().Conn()
	defer func() {
		err := conn.Close()
		if err != nil {
			h.logger.WithError(err).Error("failed to close connection")
		}
	}()

	return do(ctx, conn)
}

func (h *Handle) Client() *goredis.Client {
	return h.pool.Client(&h.ConnectionOptions)
}

func (h *Handle) NewMutex(name string) *redsync.Mutex {
	redsyncInstance := h.pool.instance(&h.ConnectionOptions).Redsync
	mutex := redsyncInstance.NewMutex(
		name,
		redsync.WithExpiry(5*time.Second),
		redsync.WithTries(5),
	)
	return mutex
}
