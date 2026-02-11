package redis

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-redsync/redsync/v4"
	goredis "github.com/redis/go-redis/v9"

	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var HandleLogger = slogutil.NewLogger("redis-handle")

type Handle struct {
	pool *Pool

	ConnectionOptions ConnectionOptions
}

func NewHandle(pool *Pool, connectionOptions ConnectionOptions) *Handle {
	return &Handle{
		pool:              pool,
		ConnectionOptions: connectionOptions,
	}
}

func (h *Handle) WithConnContext(ctx context.Context, do func(ctx context.Context, conn Redis_6_0_Cmdable) error) error {
	logger := HandleLogger.GetLogger(ctx)

	logger.With(
		slog.Int("max_open_connection", *h.ConnectionOptions.MaxOpenConnection),
		slog.Int("max_idle_connection", *h.ConnectionOptions.MaxIdleConnection),
		slog.Duration("idle_connection_timeout_seconds", h.ConnectionOptions.IdleConnectionTimeout.Duration()),
		slog.Duration("max_connection_lifetime_seconds", h.ConnectionOptions.MaxConnectionLifetime.Duration()),
	).Debug(ctx, "open redis connection")

	conn := h.Client().Conn()
	defer func() {
		err := conn.Close()
		if err != nil {
			logger.WithError(err).Error(ctx, "failed to close connection")
		}
	}()

	return do(ctx, &otelRedisConn{conn: conn})
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
