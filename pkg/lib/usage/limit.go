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

	goredis "github.com/redis/go-redis/v9"
)

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger {
	return Logger{lf.New("usage-limit")}
}

type LimitName string

type Reservation struct {
	taken  int
	name   LimitName
	config *config.UsageLimitConfig
}

var reserveLuaScript = goredis.NewScript(`
redis.replicate_commands()

local usage_limit_key = KEYS[1]
local n = tonumber(ARGV[1])
local quota = tonumber(ARGV[2])
local reset_time = tonumber(ARGV[3])

local usage = redis.pcall("GET", usage_limit_key)
if not usage then  		-- key not found
	usage = 0
elseif usage["err"] then  -- expired usage
	usage = 0
else
	usage = tonumber(usage)
end

local pass = usage + n <= quota
if pass then
	redis.call("SET", usage_limit_key, usage + n)
	redis.call("EXPIREAT", usage_limit_key, reset_time)
	usage = usage + n
end

return {pass and 1 or 0, quota - usage}
`)

func reserve(ctx context.Context, conn redis.Redis_6_0_Cmdable, key string, n int, quota int, resetTime time.Time) (bool, int64, error) {
	result, err := reserveLuaScript.Run(ctx, conn, []string{key}, n, quota, resetTime.Unix()).Slice()
	if err != nil {
		return false, 0, err
	}

	pass := result[0].(int64) == 1
	tokens := result[1].(int64)
	return pass, tokens, nil
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

func (l *Limiter) Reserve(ctx context.Context, name LimitName, config *config.UsageLimitConfig) (*Reservation, error) {
	return l.ReserveN(ctx, name, 1, config)
}

func (l *Limiter) ReserveN(ctx context.Context, name LimitName, n int, config *config.UsageLimitConfig) (*Reservation, error) {
	enabled := config.IsEnabled()
	if !enabled {
		return &Reservation{taken: 0, name: name, config: config}, nil
	}

	quota := config.GetQuota()
	key := redisLimitKey(l.AppID, name)

	pass := false
	tokens := int64(0)
	err := l.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		var err error
		pass, tokens, err = reserve(ctx, conn, key, n, quota, l.getResetTime(config))
		return err
	})
	if err != nil {
		return nil, err
	}

	l.Logger.
		WithField("key", key).
		WithField("tokens", tokens).
		WithField("pass", pass).
		Debug("check usage limit")

	if !pass {
		return nil, ErrUsageLimitExceeded(name)
	}

	return &Reservation{taken: n, name: name, config: config}, nil
}

func (l *Limiter) Cancel(ctx context.Context, r *Reservation) {
	if r == nil || r.taken == 0 {
		return
	}

	key := redisLimitKey(l.AppID, r.name)

	err := l.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		_, err := conn.IncrBy(ctx, key, -int64(r.taken)).Result()
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

	r.taken = 0
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
