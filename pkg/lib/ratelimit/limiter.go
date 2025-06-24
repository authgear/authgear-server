package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger {
	return Logger{lf.New("rate-limit")}
}

// Limiter implements rate limiting using a simple token bucket algorithm.
// Consumers take token from a bucket every operation, and tokens are refilled
// periodically.
type Limiter struct {
	Logger  Logger
	Storage Storage
	AppID   config.AppID
	Config  *config.RateLimitsFeatureConfig
}

// GetTimeToAct allows you to check what is the earliest time you can retry.
func (l *Limiter) GetTimeToAct(ctx context.Context, spec BucketSpec) (*time.Time, error) {
	_, _, timeToAct, err := l.reserveN(ctx, spec, 0)
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

// Reserve is a shortcut of ReserveN(1).
func (l *Limiter) Reserve(ctx context.Context, spec BucketSpec) (*Reservation, *FailedReservation, error) {
	return l.ReserveN(ctx, spec, 1)
}

// ReserveN is the general entry point.
// If you ever need to pass n=0, you should use GetTimeToAct() instead.
func (l *Limiter) ReserveN(ctx context.Context, spec BucketSpec, n float64) (*Reservation, *FailedReservation, error) {
	reservation, failedReservation, _, err := l.reserveN(ctx, spec, n)
	return reservation, failedReservation, err
}

func (l *Limiter) reserveN(ctx context.Context, spec BucketSpec, n float64) (*Reservation, *FailedReservation, *time.Time, error) {
	var key string
	if spec.IsGlobal {
		key = bucketKeyGlobal(spec)
	} else {
		key = bucketKeyApp(l.AppID, spec)
	}

	if l.Config.Disabled || !spec.Enabled {
		return &Reservation{
			Key:  key,
			Spec: spec,
		}, nil, nil, nil
	}

	ok, timeToAct, err := l.Storage.Update(ctx, key, spec.Period, spec.Burst, n)
	if err != nil {
		return nil, nil, nil, err
	}

	l.Logger.
		WithField("global", spec.IsGlobal).
		WithField("key", key).
		WithField("ok", ok).
		WithField("timeToAct", timeToAct).
		Debug("check rate limit")

	if ok {
		return &Reservation{
			Key:        key,
			Spec:       spec,
			TokenTaken: n,
		}, nil, &timeToAct, nil
	}

	return nil, &FailedReservation{
		key:       key,
		spec:      spec,
		timeToAct: timeToAct,
	}, &timeToAct, nil
}

func (l *Limiter) AdjustWeight(ctx context.Context, r *Reservation, weight float64) (*Reservation, *FailedReservation, error) {
	// If it was cancelled, do nothing
	if r.IsCancelled {
		return r, nil, nil
	}
	// First cancel the applied reservation
	err := l.doCancel(ctx, r)
	if err != nil {
		// Cancel failed, return error
		return nil, nil, err
	}

	// Then re-apply with the new weight
	reservation, failedReservation, _, err := l.reserveN(ctx, r.Spec, weight)
	return reservation, failedReservation, err
}

// Cancel cancels a reservation.
func (l *Limiter) Cancel(ctx context.Context, r *Reservation) {
	if r == nil || r.WasCancelPrevented {
		return
	}

	err := l.doCancel(ctx, r)
	if err != nil {
		// Errors here are non-critical and non-recoverable;
		// log and continue.
		l.Logger.WithError(err).
			WithField("global", r.Spec.IsGlobal).
			WithField("key", r.Spec.Key()).
			Warn("failed to cancel reservation")
	}
}

func (l *Limiter) doCancel(ctx context.Context, r *Reservation) error {
	if r.TokenTaken == 0 || r.IsCancelled {
		// If no token was taken or already cancelled, nothing to do
		return nil
	}
	_, _, err := l.Storage.Update(ctx, r.Key, r.Spec.Period, r.Spec.Burst, -r.TokenTaken)
	if err != nil {
		return err
	}
	r.IsCancelled = true
	return nil
}

func bucketKeyGlobal(spec BucketSpec) string {
	return fmt.Sprintf("rate-limit:%s", spec.Key())
}

func bucketKeyApp(appID config.AppID, spec BucketSpec) string {
	return fmt.Sprintf("app:%s:rate-limit:%s", appID, spec.Key())
}
