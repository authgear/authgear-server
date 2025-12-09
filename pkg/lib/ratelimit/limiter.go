package ratelimit

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var LimiterLogger = slogutil.NewLogger("rate-limit")

type LimiterEventService interface {
	DispatchEventImmediately(ctx context.Context, payload event.NonBlockingPayload) (err error)
}

// Limiter implements rate limiting using a simple token bucket algorithm.
// Consumers take token from a bucket every operation, and tokens are refilled
// periodically.
type Limiter struct {
	Database     *appdb.Handle
	Storage      Storage
	AppID        config.AppID
	Config       *config.RateLimitsFeatureConfig
	EventService LimiterEventService
}

// GetTimeToAct allows you to check what is the earliest time you can retry.
func (l *Limiter) GetTimeToAct(ctx context.Context, spec BucketSpec) (*time.Time, error) {
	_, _, timeToAct, err := l.doReserveN(ctx, spec, 0)
	if err != nil {
		return nil, err
	}
	if timeToAct != nil {
		return timeToAct, nil
	}

	zero := time.Unix(0, 0).UTC()
	return &zero, nil
}

// Allow is a shortcut of Reserve, when you do not plan to cancel the reservation.
func (l *Limiter) Allow(ctx context.Context, spec BucketSpec) (*FailedReservation, error) {
	_, failedReservation, err := l.Reserve(ctx, spec)
	return failedReservation, err
}

// Reserve is reserveN(weight).
// weight default is 1, but it can be modified by user.
func (l *Limiter) Reserve(ctx context.Context, spec BucketSpec) (*Reservation, *FailedReservation, error) {
	weight := spec.RateLimit.ResolveWeight(ctx)
	return l.reserveN(ctx, spec, weight)
}

// reserveN is the general entry point.
// If you ever need to pass n=0, you should use GetTimeToAct() instead.
func (l *Limiter) reserveN(ctx context.Context, spec BucketSpec, n float64) (*Reservation, *FailedReservation, error) {
	reservation, failedReservation, _, err := l.doReserveN(ctx, spec, n)
	return reservation, failedReservation, err
}

func (l *Limiter) doReserveN(ctx context.Context, spec BucketSpec, n float64) (*Reservation, *FailedReservation, *time.Time, error) {
	logger := LimiterLogger.GetLogger(ctx)
	var key string
	if spec.IsGlobal {
		key = bucketKeyGlobal(spec)
	} else {
		key = bucketKeyApp(l.AppID, spec)
	}

	if l.Config.Disabled || !spec.Enabled {
		return &Reservation{
			key:  key,
			spec: spec,
		}, nil, nil, nil
	}

	ok, timeToAct, err := l.Storage.Update(ctx, key, spec.Period, spec.Burst, n)
	if err != nil {
		return nil, nil, nil, err
	}

	logger.Debug(ctx, "check rate limit",
		slog.Bool("global", spec.IsGlobal),
		slog.String("key", key),
		slog.String("bucket", string(spec.Name)),
		slog.String("ratelimit", string(spec.RateLimit)),
		slog.Bool("ok", ok),
		slog.Time("timeToAct", timeToAct),
	)

	if ok {
		return &Reservation{
			key:        key,
			spec:       spec,
			tokenTaken: n,
		}, nil, &timeToAct, nil
	}

	logger.WithSkipStackTrace().WithSkipLogging().Error(ctx, "rate limited",
		slog.Bool("global", spec.IsGlobal),
		slog.String("bucket", string(spec.Name)),
		slog.String("ratelimit", string(spec.RateLimit)),
		slog.String("key", key),
		slog.Bool("ok", ok),
		slog.Bool("ratelimit_logging", true),
		slog.Time("timeToAct", timeToAct),
	)

	if spec.RateLimit != "" {
		// Create audit log only if the rate limit is part of the public api
		ev := nonblocking.RateLimitBlockedEventPayload{
			RateLimit: string(spec.RateLimit),
			Bucket:    string(spec.Name),
		}
		var logErr error
		// Limiter might be used outside transaction, so we need to check if there is an open transaction first.
		if l.Database.IsInTx(ctx) {
			logErr = l.EventService.DispatchEventImmediately(ctx, &ev)
		} else {
			logErr = l.Database.WithTx(ctx, func(ctx context.Context) error {
				return l.EventService.DispatchEventImmediately(ctx, &ev)
			})
		}
		if logErr != nil {
			// Log the error and continue to ensure the api returns a RateLimited error
			logger.WithError(logErr).Error(ctx, "failed to dispatch rate_limit.blocked")
		}
	}

	return nil, &FailedReservation{
		key:       key,
		spec:      spec,
		timeToAct: timeToAct,
	}, &timeToAct, nil
}

// Cancel cancels a reservation.
func (l *Limiter) Cancel(ctx context.Context, r *Reservation) {
	logger := LimiterLogger.GetLogger(ctx)
	if r == nil || r.wasCancelPrevented || r.tokenTaken == 0 {
		return
	}

	_, _, err := l.Storage.Update(ctx, r.key, r.spec.Period, r.spec.Burst, -r.tokenTaken)
	if err != nil {
		// Errors here are non-critical and non-recoverable;
		// log and continue.
		logger.WithError(err).Warn(ctx, "failed to cancel reservation",
			slog.Bool("global", r.spec.IsGlobal),
			slog.String("key", r.spec.Key()),
		)
	}
}

func bucketKeyGlobal(spec BucketSpec) string {
	return fmt.Sprintf("rate-limit:%s", spec.Key())
}

func bucketKeyApp(appID config.AppID, spec BucketSpec) string {
	return fmt.Sprintf("app:%s:rate-limit:%s", appID, spec.Key())
}
