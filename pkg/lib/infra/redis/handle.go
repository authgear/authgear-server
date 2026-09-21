package redis

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/go-redsync/redsync/v4"
	goredis "github.com/redis/go-redis/v9"

	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

// ErrMutexNotAcquired is returned by WithMutexExpiry specifically when the
// mutex is already held elsewhere (redsync.ErrTaken / redsync.ErrNodeTaken)
// -- contention a caller polling on an interval expects and should treat as
// routine, not an error. It is deliberately not either of those two
// concrete types so callers have one sentinel to check against regardless
// of which fired. Any other failure (a Redis connectivity/auth error, for
// example) is a real error and is returned as-is, uncollapsed, so it
// reaches the caller's normal error handling instead of being
// indistinguishable from routine contention.
var ErrMutexNotAcquired = errors.New("redis: mutex not acquired")

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

func (h *Handle) WithMutex(ctx context.Context, name string, do func() error) error {
	mutex := h.NewMutex(name)
	if err := mutex.LockContext(ctx); err != nil {
		return err
	}
	unlockCtx := context.WithoutCancel(ctx)
	defer func() {
		_, _ = mutex.UnlockContext(unlockCtx)
	}()
	return do()
}

// WithMutexExpiry is like WithMutex, but for a critical section whose
// duration is not knowable in advance -- for example a batch delivery
// across every app and stream in one tick -- so the caller picks the lock
// expiry instead of the fixed 5s/5-tries NewMutex uses. It makes exactly
// one attempt: a caller polling on an interval wants to skip a tick it
// cannot get the lock for, not block waiting on it.
func (h *Handle) WithMutexExpiry(ctx context.Context, name string, expiry time.Duration, do func() error) error {
	redsyncInstance := h.pool.instance(&h.ConnectionOptions).Redsync
	mutex := redsyncInstance.NewMutex(
		name,
		redsync.WithExpiry(expiry),
		redsync.WithTries(1),
	)
	if err := mutex.LockContext(ctx); err != nil {
		var errTaken *redsync.ErrTaken
		var errNodeTaken *redsync.ErrNodeTaken
		if errors.As(err, &errTaken) || errors.As(err, &errNodeTaken) {
			return ErrMutexNotAcquired
		}
		return err
	}
	unlockCtx := context.WithoutCancel(ctx)
	defer func() {
		_, _ = mutex.UnlockContext(unlockCtx)
	}()
	return do()
}
