package ratelimit

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var LimiterGlobalLogger = slogutil.NewLogger("limiter-global")

type LimiterGlobal struct {
	Storage Storage
}

// GetTimeToAct allows you to check what is the earliest time you can retry.
func (l *LimiterGlobal) GetTimeToAct(ctx context.Context, spec BucketSpec) (*time.Time, error) {
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
func (l *LimiterGlobal) Allow(ctx context.Context, spec BucketSpec) (*FailedReservation, error) {
	_, failedReservation, err := l.Reserve(ctx, spec)
	return failedReservation, err
}

// Reserve is reserveN(weight).
// weight default is 1, but it can be modified by user.
func (l *LimiterGlobal) Reserve(ctx context.Context, spec BucketSpec) (*Reservation, *FailedReservation, error) {
	weight := spec.RateLimitGroup.ResolveWeight(ctx)
	return l.reserveN(ctx, spec, weight)
}

// reserveN is the general entry point.
// If you ever need to pass n=0, you should use GetTimeToAct() instead.
func (l *LimiterGlobal) reserveN(ctx context.Context, spec BucketSpec, n float64) (*Reservation, *FailedReservation, error) {
	reservation, failedReservation, _, err := l.doReserveN(ctx, spec, n)
	return reservation, failedReservation, err
}

func (l *LimiterGlobal) doReserveN(ctx context.Context, spec BucketSpec, n float64) (*Reservation, *FailedReservation, *time.Time, error) {
	logger := LimiterGlobalLogger.GetLogger(ctx)
	key := bucketKeyGlobal(spec)

	if !spec.IsGlobal {
		panic(errors.New("ratelimit: must be global limit"))
	}

	if !spec.Enabled {
		return &Reservation{
			key:  key,
			spec: spec,
		}, nil, nil, nil
	}

	ok, timeToAct, err := l.Storage.Update(ctx, key, spec.Period, spec.Burst, n)
	if err != nil {
		return nil, nil, nil, nil
	}

	logger.With(
		slog.String("key", spec.Key()),
		slog.String("bucket", string(spec.Name)),
		slog.String("ratelimit", string(spec.RateLimitGroup)),
		slog.Bool("ok", ok),
		slog.Time("timeToAct", timeToAct),
	).Debug(ctx, "check global rate limit")

	if ok {
		return &Reservation{
			spec:       spec,
			key:        key,
			tokenTaken: n,
		}, nil, &timeToAct, nil
	}

	return nil, &FailedReservation{
		spec:      spec,
		key:       key,
		timeToAct: timeToAct,
	}, &timeToAct, nil
}

// Cancel cancels a reservation.
func (l *LimiterGlobal) Cancel(ctx context.Context, r *Reservation) {
	logger := LimiterGlobalLogger.GetLogger(ctx)

	if r == nil || r.wasCancelPrevented || r.tokenTaken == 0 {
		return
	}

	_, _, err := l.Storage.Update(ctx, r.key, r.spec.Period, r.spec.Burst, -r.tokenTaken)
	if err != nil {
		// Errors here are non-critical and non-recoverable;
		// log and continue.
		logger.WithError(err).With(
			slog.Bool("global", r.spec.IsGlobal),
			slog.String("key", r.spec.Key()),
		).Warn(ctx, "failed to cancel reservation")
	}
}
