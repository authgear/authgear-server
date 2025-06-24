package ratelimit

import (
	"context"
	"errors"
	"time"
)

type LimiterGlobal struct {
	Logger  Logger
	Storage Storage
}

// GetTimeToAct allows you to check what is the earliest time you can retry.
func (l *LimiterGlobal) GetTimeToAct(ctx context.Context, spec BucketSpec) (*time.Time, error) {
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
func (l *LimiterGlobal) Allow(ctx context.Context, spec BucketSpec) (*FailedReservation, error) {
	_, failedReservation, err := l.Reserve(ctx, spec)
	return failedReservation, err
}

// Reserve is a shortcut of ReserveN(1).
func (l *LimiterGlobal) Reserve(ctx context.Context, spec BucketSpec) (*Reservation, *FailedReservation, error) {
	return l.ReserveN(ctx, spec, 1)
}

// ReserveN is the general entry point.
// If you ever need to pass n=0, you should use GetTimeToAct() instead.
func (l *LimiterGlobal) ReserveN(ctx context.Context, spec BucketSpec, n float64) (*Reservation, *FailedReservation, error) {
	reservation, failedReservation, _, err := l.reserveN(ctx, spec, n)
	return reservation, failedReservation, err
}

func (l *LimiterGlobal) reserveN(ctx context.Context, spec BucketSpec, n float64) (*Reservation, *FailedReservation, *time.Time, error) {
	key := bucketKeyGlobal(spec)

	if !spec.IsGlobal {
		panic(errors.New("ratelimit: must be global limit"))
	}

	if !spec.Enabled {
		return &Reservation{
			Key:  key,
			Spec: spec,
		}, nil, nil, nil
	}

	ok, timeToAct, err := l.Storage.Update(ctx, key, spec.Period, spec.Burst, n)
	if err != nil {
		return nil, nil, nil, nil
	}

	l.Logger.
		WithField("key", spec.Key()).
		WithField("ok", ok).
		WithField("timeToAct", timeToAct).
		Debug("check global rate limit")

	if ok {
		return &Reservation{
			Spec:       spec,
			Key:        key,
			TokenTaken: n,
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
	if r == nil || r.WasCancelPrevented || r.TokenTaken == 0 {
		return
	}

	_, _, err := l.Storage.Update(ctx, r.Key, r.Spec.Period, r.Spec.Burst, -r.TokenTaken)
	if err != nil {
		// Errors here are non-critical and non-recoverable;
		// log and continue.
		l.Logger.WithError(err).
			WithField("global", r.Spec.IsGlobal).
			WithField("key", r.Spec.Key()).
			Warn("failed to cancel reservation")
	}
}
