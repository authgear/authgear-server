package usage

import (
	"context"
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger {
	return Logger{lf.New("usage-limit")}
}

type LimitName string

type Reservation struct {
	taken  bool
	name   LimitName
	config *config.UsageLimitConfig
}

type Limiter struct {
	Logger Logger
	Clock  clock.Clock
	AppID  config.AppID
	Redis  *appredis.Handle
}

func (l *Limiter) getResetTime(c *config.UsageLimitConfig) time.Time {
	return ComputeResetTime(l.Clock.NowUTC(), c.Period)
}

func (l *Limiter) Reserve(name LimitName, config *config.UsageLimitConfig) (*Reservation, error) {
	enabled := config.Enabled != nil && *config.Enabled
	if !enabled {
		return &Reservation{taken: false, name: name, config: config}, nil
	}

	quota := config.Quota
	key := redisLimitKey(l.AppID, name)

	tokens := int64(0)
	err := l.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
		ctx := context.Background()
		usage, err := conn.IncrBy(ctx, key, 1).Result()
		if err != nil {
			return err
		}

		tokens = int64(quota) - usage

		resetTime := l.getResetTime(config)
		// Ignore error
		_, _ = conn.PExpireAt(ctx, key, resetTime).Result()

		return nil
	})
	if err != nil {
		return nil, err
	}

	pass := tokens >= 0
	l.Logger.
		WithField("key", key).
		WithField("tokens", tokens).
		WithField("pass", pass).
		Debug("check usage limit")

	if !pass {
		return nil, ErrUsageLimitExceeded(name)
	}

	return &Reservation{taken: true, name: name, config: config}, nil
}

func (l *Limiter) Cancel(r *Reservation) {
	if !r.taken {
		return
	}

	key := redisLimitKey(l.AppID, r.name)

	err := l.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
		ctx := context.Background()
		_, err := conn.IncrBy(ctx, key, -1).Result()
		if err != nil {
			return err
		}

		resetTime := l.getResetTime(r.config)
		// Ignore error
		_, _ = conn.PExpireAt(ctx, key, resetTime).Result()

		return nil
	})

	if err != nil {
		// Errors here are non-critical and non-recoverable;
		// log and continue.
		l.Logger.WithError(err).
			WithField("key", key).
			Warn("failed to cancel reservation")
	}
}

func redisLimitKey(appID config.AppID, name LimitName) string {
	return fmt.Sprintf("app:%s:usage-limit:%s", appID, name)
}

func ComputeResetTime(now time.Time, period config.UsageLimitPeriod) time.Time {
	switch period {
	case config.UsageLimitPeriodDay:
		return now.Truncate(24*time.Hour).AddDate(0, 0, 1)
	case config.UsageLimitPeriodMonth:
		return now.Truncate(24*time.Hour).AddDate(0, 1, -now.Day()+1)
	default:
		panic("usage: unknown usage limit period: " + period)
	}
}
