package redis_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redsync/redsync/v4"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
)

func newTestHandle(t *testing.T, redisURL string) *redis.Handle {
	t.Helper()
	maxOpen := 10
	maxIdle := 10
	idleTimeout := config.DurationSeconds(60)
	maxLifetime := config.DurationSeconds(60)

	pool := redis.NewPool()
	t.Cleanup(func() { _ = pool.Close() })

	return redis.NewHandle(pool, redis.ConnectionOptions{
		RedisURL:              redisURL,
		MaxOpenConnection:     &maxOpen,
		MaxIdleConnection:     &maxIdle,
		IdleConnectionTimeout: &idleTimeout,
		MaxConnectionLifetime: &maxLifetime,
	})
}

func TestHandleWithMutexExpiry(t *testing.T) {
	ctx := context.Background()

	Convey("WithMutexExpiry", t, func() {
		Convey("contention (mutex already held) collapses to ErrMutexNotAcquired", func() {
			s := miniredis.RunT(t)
			h := newTestHandle(t, "redis://"+s.Addr())

			// Hold the lock directly, bypassing WithMutexExpiry, so the
			// second attempt genuinely contends rather than racing itself.
			holder := h.NewMutex("test-mutex")
			err := holder.LockContext(ctx)
			So(err, ShouldBeNil)

			called := false
			err = h.WithMutexExpiry(ctx, "test-mutex", time.Second, func() error {
				called = true
				return nil
			})
			So(called, ShouldBeFalse)
			So(errors.Is(err, redis.ErrMutexNotAcquired), ShouldBeTrue)
		})

		Convey("a Redis connectivity failure is not collapsed into ErrMutexNotAcquired", func() {
			// Port 1 is not listening, so the dial fails immediately
			// rather than hanging until a timeout.
			h := newTestHandle(t, "redis://127.0.0.1:1")

			called := false
			err := h.WithMutexExpiry(ctx, "test-mutex", time.Second, func() error {
				called = true
				return nil
			})
			So(called, ShouldBeFalse)
			So(err, ShouldNotBeNil)
			So(errors.Is(err, redis.ErrMutexNotAcquired), ShouldBeFalse)

			var errTaken *redsync.ErrTaken
			var errNodeTaken *redsync.ErrNodeTaken
			So(errors.As(err, &errTaken), ShouldBeFalse)
			So(errors.As(err, &errNodeTaken), ShouldBeFalse)
		})

		Convey("do() runs and its error propagates when the lock is acquired", func() {
			s := miniredis.RunT(t)
			h := newTestHandle(t, "redis://"+s.Addr())

			sentinel := errors.New("boom")
			err := h.WithMutexExpiry(ctx, "test-mutex", time.Second, func() error {
				return sentinel
			})
			So(errors.Is(err, sentinel), ShouldBeTrue)
		})

		Convey("the lock is renewed while do() runs, surviving past the original expiry", func() {
			s := miniredis.RunT(t)
			h := newTestHandle(t, "redis://"+s.Addr())

			expiry := 300 * time.Millisecond
			doStarted := make(chan struct{})
			doFinished := make(chan struct{})
			var doErr error
			go func() {
				doErr = h.WithMutexExpiry(ctx, "test-mutex", expiry, func() error {
					close(doStarted)
					// Well past expiry: without renewal the lock would
					// already be gone from Redis partway through this.
					time.Sleep(3 * expiry)
					return nil
				})
				close(doFinished)
			}()

			<-doStarted
			// Past the original expiry, but do() is still running.
			time.Sleep(2 * expiry)

			// If the lock were not being renewed, it would have expired
			// in Redis by now and this would succeed instead.
			contendErr := h.WithMutexExpiry(ctx, "test-mutex", time.Second, func() error {
				return nil
			})

			<-doFinished
			So(doErr, ShouldBeNil)
			So(errors.Is(contendErr, redis.ErrMutexNotAcquired), ShouldBeTrue)
		})
	})
}
