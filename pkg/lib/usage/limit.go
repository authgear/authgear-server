package usage

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var logger = slogutil.NewLogger("usage-limit")

type LimitName string

type Reservation struct {
	taken   int
	name    model.UsageName
	config  *config.Deprecated_UsageLimitConfig
	periods []model.UsageLimitPeriod
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
	Clock           clock.Clock
	AppID           config.AppID
	Redis           *appredis.Handle
	EffectiveConfig *config.Config
}

func (l *Limiter) getResetTime(c *config.Deprecated_UsageLimitConfig) time.Time {
	return ComputeResetTime(l.Clock.NowUTC(), model.UsageLimitPeriod(c.Period))
}

func (l *Limiter) Reserve(ctx context.Context, name model.UsageName, n int) (*Reservation, error) {
	logger := logger.GetLogger(ctx)
	config := l.effectiveDeprecatedUsageLimit(name)
	enabled := config.IsEnabled()
	if !enabled {
		return &Reservation{taken: 0, name: name, config: config}, nil
	}

	quota := config.GetQuota()
	configuredPeriod := model.UsageLimitPeriod(config.Period)
	key := redisLimitKey(l.AppID, name, configuredPeriod)

	pass := false
	tokens := int64(0)
	err := l.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		var err error
		pass, tokens, err = reserve(ctx, conn, key, n, quota, ComputeResetTime(l.Clock.NowUTC(), configuredPeriod))
		if err != nil {
			return err
		}
		if !pass {
			return nil
		}
		for _, period := range usagePeriods() {
			if period == configuredPeriod {
				continue
			}
			otherKey := redisLimitKey(l.AppID, name, period)
			if _, err = conn.IncrBy(ctx, otherKey, int64(n)).Result(); err != nil {
				return err
			}
			if _, err = conn.PExpireAt(ctx, otherKey, ComputeResetTime(l.Clock.NowUTC(), period)).Result(); err != nil {
				return err
			}
		}
		return err
	})
	if err != nil {
		return nil, err
	}

	logger.With(
		slog.String("key", key),
		slog.Int64("tokens", tokens),
		slog.Bool("pass", pass),
	).Debug(ctx, "check usage limit")

	if !pass {
		return nil, ErrUsageLimitExceeded(name, model.UsageLimitPeriod(config.Period))
	}

	return &Reservation{taken: n, name: name, config: config, periods: usagePeriods()}, nil
}

func (l *Limiter) Cancel(ctx context.Context, r *Reservation) {
	logger := logger.GetLogger(ctx)
	if r == nil || r.taken == 0 {
		return
	}

	err := l.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		for _, period := range r.periods {
			key := redisLimitKey(l.AppID, r.name, period)
			_, err := conn.IncrBy(ctx, key, -int64(r.taken)).Result()
			if err != nil {
				return err
			}

			resetTime := ComputeResetTime(l.Clock.NowUTC(), period)
			// Ignore error
			_, _ = conn.PExpireAt(ctx, key, resetTime).Result()
		}

		return nil
	})

	if err != nil {
		// Errors here are non-critical and non-recoverable;
		// log and continue.
		logger.WithError(err).Warn(ctx, "failed to cancel reservation")
	}

	r.taken = 0
}

func redisLimitKey(appID config.AppID, name model.UsageName, period model.UsageLimitPeriod) string {
	legacyName := legacyLimitName(name)
	if period == model.UsageLimitPeriodMonth {
		return fmt.Sprintf("app:%s:usage-limit:%s", appID, legacyName)
	}
	return fmt.Sprintf("app:%s:usage-limit:%s:%s", appID, legacyName, period)
}

func (l *Limiter) effectiveDeprecatedUsageLimit(name model.UsageName) *config.Deprecated_UsageLimitConfig {
	if l == nil || l.EffectiveConfig == nil || l.EffectiveConfig.FeatureConfig == nil {
		return nil
	}

	featureConfig := l.EffectiveConfig.FeatureConfig
	switch name {
	case model.UsageNameEmail:
		if featureConfig.Messaging != nil {
			return featureConfig.Messaging.EmailUsage
		}
	case model.UsageNameSMS:
		if featureConfig.Messaging != nil {
			return featureConfig.Messaging.SMSUsage
		}
	case model.UsageNameWhatsapp:
		if featureConfig.Messaging != nil {
			return featureConfig.Messaging.WhatsappUsage
		}
	case model.UsageNameUserImport:
		if featureConfig.AdminAPI != nil {
			return featureConfig.AdminAPI.UserImportUsage
		}
	case model.UsageNameUserExport:
		if featureConfig.AdminAPI != nil {
			return featureConfig.AdminAPI.UserExportUsage
		}
	}

	return nil
}

func legacyLimitName(name model.UsageName) LimitName {
	switch name {
	case model.UsageNameEmail:
		return LimitNameEmail
	case model.UsageNameSMS:
		return LimitNameSMS
	case model.UsageNameWhatsapp:
		return LimitNameWhatsapp
	case model.UsageNameUserImport:
		return LimitNameUserImport
	case model.UsageNameUserExport:
		return LimitNameUserExport
	default:
		panic("usage: unknown usage name: " + string(name))
	}
}

func ComputeResetTime(now time.Time, period model.UsageLimitPeriod) time.Time {
	switch period {
	case model.UsageLimitPeriodDay:
		return now.Truncate(24*time.Hour).AddDate(0, 0, 1)
	case model.UsageLimitPeriodMonth:
		return now.Truncate(24*time.Hour).AddDate(0, 1, -now.Day()+1)
	default:
		panic("usage: unknown usage limit period: " + period)
	}
}

func usagePeriods() []model.UsageLimitPeriod {
	return []model.UsageLimitPeriod{
		model.UsageLimitPeriodDay,
		model.UsageLimitPeriodMonth,
	}
}
