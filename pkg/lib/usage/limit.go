package usage

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	goredis "github.com/redis/go-redis/v9"

	apievent "github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var logger = slogutil.NewLogger("usage-limit")

type LimitName string

type EffectiveUsageLimit struct {
	Name   model.UsageName
	Quota  int
	Period model.UsageLimitPeriod
	Action model.UsageLimitAction
}

type Reservation struct {
	name    model.UsageName
	taken   int
	results []periodReservationResult
}

type periodReservationResult struct {
	Period    model.UsageLimitPeriod
	Limits    []EffectiveUsageLimit
	Key       string
	ResetTime time.Time
	Quota     int
	Pass      bool
	Before    int
	After     int
	Taken     int
}

var reserveLuaScript = goredis.NewScript(`
redis.replicate_commands()

local usage_limit_key = KEYS[1]
local n = tonumber(ARGV[1])
local reset_time = tonumber(ARGV[2])
local quota = tonumber(ARGV[3])

local usage = redis.pcall("GET", usage_limit_key)
if not usage then
	usage = 0
elseif usage["err"] then
	usage = 0
else
	usage = tonumber(usage)
end

local usage_before = usage
local usage_after = usage

if quota >= 0 and usage_before + n > quota then
	return {0, usage_before, usage_after}
end

usage_after = usage_before + n
redis.call("SET", usage_limit_key, usage_after)
redis.call("EXPIREAT", usage_limit_key, reset_time)

return {1, usage_before, usage_after}
`)

type Limiter struct {
	Clock                  clock.Clock
	Database               *appdb.Handle
	AppID                  config.AppID
	Redis                  *appredis.Handle
	EffectiveConfig        *config.Config
	EventService           EventService
	UsageAlertEmailService UsageAlertEmailService
}

type EventService interface {
	DispatchEventImmediately(ctx context.Context, payload apievent.NonBlockingPayload) error
}

func (l *Limiter) Reserve(ctx context.Context, name model.UsageName, n int) (*Reservation, error) {
	logger := logger.GetLogger(ctx)
	limits := l.effectiveUsageLimits(name)
	if len(limits) == 0 {
		return &Reservation{name: name}, nil
	}

	reservation := &Reservation{
		name:  name,
		taken: n,
	}

	for _, period := range l.usagePeriods() {
		periodLimits := l.limitsForPeriod(limits, period)
		result, err := l.reservePeriod(ctx, name, period, n, periodLimits)
		if err != nil {
			l.Cancel(ctx, reservation)
			return nil, err
		}
		if !result.Pass {
			logger.With(
				slog.String("usage_name", string(name)),
				slog.String("period", string(result.Period)),
				slog.Int("taken", result.Taken),
				slog.Int("quota", result.Quota),
				slog.Int("before", result.Before),
				slog.Int("after", result.After),
			).Warn(ctx, "usage reservation blocked")
			if err := l.rollbackPeriodResults(ctx, reservation.results); err != nil {
				logger.WithError(err).Warn(ctx, "failed to rollback usage reservation")
			}
			_ = l.evaluateUsageTriggers(ctx, name, result.Period, result.Before, result.After, result.Limits)
			return nil, ErrUsageLimitExceeded(name, result.Period)
		}
		reservation.results = append(reservation.results, *result)
	}

	for _, result := range reservation.results {
		_ = l.evaluateUsageTriggers(ctx, name, result.Period, result.Before, result.After, result.Limits)
	}

	return reservation, nil
}

func (l *Limiter) reservePeriod(ctx context.Context, name model.UsageName, period model.UsageLimitPeriod, n int, limits []EffectiveUsageLimit) (*periodReservationResult, error) {
	resetTime := ComputeResetTime(l.Clock.NowUTC(), period)
	key := l.redisLimitKey(name, period)
	blockQuota, hasBlockQuota := l.minBlockQuota(limits)

	result := &periodReservationResult{
		Period:    period,
		Limits:    limits,
		Key:       key,
		ResetTime: resetTime,
		Taken:     n,
	}

	if hasBlockQuota {
		result.Quota = blockQuota
		pass, before, after, err := l.reserveWithQuota(ctx, key, n, blockQuota, resetTime)
		if err != nil {
			return nil, err
		}
		result.Pass = pass
		result.Before = before
		result.After = after
		return result, nil
	}

	before, after, err := l.incrementWithoutQuota(ctx, key, n, resetTime)
	if err != nil {
		return nil, err
	}
	result.Pass = true
	result.Before = before
	result.After = after
	return result, nil
}

func (l *Limiter) reserveWithQuota(ctx context.Context, key string, n int, quota int, resetTime time.Time) (pass bool, before int, after int, err error) {
	err = l.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		var innerErr error
		pass, before, after, innerErr = runReserveScript(ctx, conn, key, n, resetTime, quota)
		return innerErr
	})
	return
}

func (l *Limiter) incrementWithoutQuota(ctx context.Context, key string, n int, resetTime time.Time) (before int, after int, err error) {
	var pass bool
	err = l.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		var innerErr error
		pass, before, after, innerErr = runReserveScript(ctx, conn, key, n, resetTime, -1)
		return innerErr
	})
	if err != nil {
		return 0, 0, err
	}
	if !pass {
		panic("usage: incrementWithoutQuota unexpectedly failed")
	}
	return before, after, nil
}

func (l *Limiter) evaluateUsageTriggers(ctx context.Context, name model.UsageName, period model.UsageLimitPeriod, before, after int, limits []EffectiveUsageLimit) error {
	for _, limit := range crossedUsageLimits(before, after, limits) {
		if err := l.maybeDispatchUsageAlert(ctx, limit, after); err != nil {
			return err
		}
	}
	return nil
}

func (l *Limiter) Cancel(ctx context.Context, r *Reservation) {
	logger := logger.GetLogger(ctx)
	if r == nil || r.taken == 0 || len(r.results) == 0 {
		return
	}

	if err := l.rollbackPeriodResults(ctx, r.results); err != nil {
		logger.WithError(err).Warn(ctx, "failed to cancel reservation")
	}

	r.taken = 0
	r.results = nil
}

func (l *Limiter) rollbackPeriodResults(ctx context.Context, results []periodReservationResult) error {
	if len(results) == 0 {
		return nil
	}

	return l.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		for i := len(results) - 1; i >= 0; i-- {
			result := results[i]
			if _, err := conn.IncrBy(ctx, result.Key, -int64(result.Taken)).Result(); err != nil {
				return err
			}
			_, _ = conn.PExpireAt(ctx, result.Key, result.ResetTime).Result()
		}
		return nil
	})
}

func (l *Limiter) effectiveUsageLimits(name model.UsageName) []EffectiveUsageLimit {
	var limits []EffectiveUsageLimit

	if l.EffectiveConfig.FeatureConfig.Usage != nil && l.EffectiveConfig.FeatureConfig.Usage.Limits != nil {
		for _, limit := range l.EffectiveConfig.FeatureConfig.Usage.Limits.Limits(name) {
			limits = append(limits, EffectiveUsageLimit{
				Name:   name,
				Quota:  limit.Quota,
				Period: limit.Period,
				Action: limit.Action,
			})
		}
	}

	if l.EffectiveConfig.AppConfig.Usage != nil && l.EffectiveConfig.AppConfig.Usage.Limits != nil {
		for _, limit := range l.EffectiveConfig.AppConfig.Usage.Limits.Limits(name) {
			limits = append(limits, EffectiveUsageLimit{
				Name:   name,
				Quota:  limit.Quota,
				Period: limit.Period,
				Action: limit.Action,
			})
		}
	}

	if len(limits) > 0 {
		return limits
	}

	return nil
}

func (l *Limiter) usagePeriods() []model.UsageLimitPeriod {
	return []model.UsageLimitPeriod{
		model.UsageLimitPeriodDay,
		model.UsageLimitPeriodMonth,
	}
}

func (l *Limiter) limitsForPeriod(limits []EffectiveUsageLimit, period model.UsageLimitPeriod) []EffectiveUsageLimit {
	var filtered []EffectiveUsageLimit
	for _, limit := range limits {
		if limit.Period == period {
			filtered = append(filtered, limit)
		}
	}
	return filtered
}

func (l *Limiter) usageHookURLs(name model.UsageName) []string {
	if l.EffectiveConfig.FeatureConfig.Usage == nil {
		return nil
	}

	var urls []string
	for _, hook := range l.EffectiveConfig.FeatureConfig.Usage.Hooks {
		if hook.Match == "*" || hook.Match == string(name) {
			urls = append(urls, hook.URL)
		}
	}
	return urls
}

func (l *Limiter) usageAlertRecipients(name model.UsageName) []string {
	if l.EffectiveConfig.AppConfig.Usage == nil {
		return nil
	}

	seen := map[string]struct{}{}
	var recipients []string
	for _, alert := range l.EffectiveConfig.AppConfig.Usage.Alerts {
		if alert.Type != "email" {
			continue
		}
		if alert.Match != "*" && alert.Match != string(name) {
			continue
		}
		if _, ok := seen[alert.Email]; ok {
			continue
		}
		seen[alert.Email] = struct{}{}
		recipients = append(recipients, alert.Email)
	}
	return recipients
}

func (l *Limiter) makeUsageAlertTriggeredPayload(ctx context.Context, limit EffectiveUsageLimit, currentValue int) *nonblocking.UsageAlertTriggeredEventPayload {
	return &nonblocking.UsageAlertTriggeredEventPayload{
		Usage: nonblocking.UsageAlertPayload{
			Name:         limit.Name,
			Action:       limit.Action,
			Period:       limit.Period,
			Quota:        limit.Quota,
			CurrentValue: currentValue,
			PlanName:     usagePlanName(ctx),
		},
		HookURLs: l.usageHookURLs(limit.Name),
	}
}

func (l *Limiter) maybeDispatchUsageAlert(ctx context.Context, limit EffectiveUsageLimit, currentValue int) error {
	logger := logger.GetLogger(ctx)
	payload := l.makeUsageAlertTriggeredPayload(ctx, limit, currentValue)
	if err := l.dispatchEventImmediately(ctx, payload); err != nil {
		logger.WithError(err).Warn(ctx, "failed to dispatch usage alert event")
	}
	if l.UsageAlertEmailService != nil {
		recipients := l.usageAlertRecipients(limit.Name)
		if err := l.UsageAlertEmailService.Send(ctx, recipients, payload); err != nil {
			logger.WithError(err).Warn(ctx, "failed to send usage alert email")
		}
	}
	return nil
}

func usagePlanName(ctx context.Context) string {
	appCtx, ok := config.GetAppContext(ctx)
	if !ok || appCtx == nil {
		return ""
	}
	return appCtx.PlanName
}

func (l *Limiter) dispatchEventImmediately(ctx context.Context, payload apievent.NonBlockingPayload) error {
	if l.Database.IsInTx(ctx) {
		return l.EventService.DispatchEventImmediately(ctx, payload)
	}

	return l.Database.ReadOnly(ctx, func(ctx context.Context) error {
		return l.EventService.DispatchEventImmediately(ctx, payload)
	})
}

func (l *Limiter) minBlockQuota(limits []EffectiveUsageLimit) (int, bool) {
	minQuota := 0
	found := false
	for _, limit := range limits {
		if limit.Action != model.UsageLimitActionBlock {
			continue
		}
		if !found || limit.Quota < minQuota {
			minQuota = limit.Quota
			found = true
		}
	}
	return minQuota, found
}

func (l *Limiter) redisLimitKey(name model.UsageName, period model.UsageLimitPeriod) string {
	legacyName := legacyLimitName(name)
	if period == model.UsageLimitPeriodMonth {
		return fmt.Sprintf("app:%s:usage-limit:%s", l.AppID, legacyName)
	}
	return fmt.Sprintf("app:%s:usage-limit:%s:%s", l.AppID, legacyName, period)
}

func runReserveScript(ctx context.Context, conn redis.Redis_6_0_Cmdable, key string, n int, resetTime time.Time, quota int) (pass bool, before int, after int, err error) {
	result, err := reserveLuaScript.Run(ctx, conn, []string{key}, n, resetTime.Unix(), quota).Slice()
	if err != nil {
		return false, 0, 0, err
	}

	pass = result[0].(int64) == 1
	before = int(result[1].(int64))
	after = int(result[2].(int64))
	return pass, before, after, nil
}

func crossedUsageLimits(before int, after int, limits []EffectiveUsageLimit) []EffectiveUsageLimit {
	var crossed []EffectiveUsageLimit
	for _, limit := range limits {
		if before < limit.Quota && after >= limit.Quota {
			crossed = append(crossed, limit)
		}
	}
	return crossed
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
