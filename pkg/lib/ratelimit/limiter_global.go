package ratelimit

import "errors"

type LimiterGlobal struct {
	Logger  Logger
	Storage Storage
}

func (l *LimiterGlobal) Allow(spec BucketSpec) (*FailedReservation, error) {
	_, failedReservation, err := l.Reserve(spec)
	return failedReservation, err
}

func (l *LimiterGlobal) Reserve(spec BucketSpec) (*Reservation, *FailedReservation, error) {
	return l.ReserveN(spec, 1)
}

func (l *LimiterGlobal) ReserveN(spec BucketSpec, n int) (*Reservation, *FailedReservation, error) {
	key := bucketKeyGlobal(spec)

	if !spec.IsGlobal {
		panic(errors.New("ratelimit: must be global limit"))
	}

	if !spec.Enabled {
		return &Reservation{
			key:  key,
			spec: spec,
		}, nil, nil
	}

	ok, timeToAct, err := l.Storage.Update(key, spec.Period, spec.Burst, n)
	if err != nil {
		return nil, nil, nil
	}

	l.Logger.
		WithField("key", spec.Key()).
		WithField("ok", ok).
		WithField("timeToAct", timeToAct).
		Debug("check global rate limit")

	if ok {
		return &Reservation{
			spec:       spec,
			key:        key,
			tokenTaken: n,
		}, nil, nil
	}

	return nil, &FailedReservation{
		spec:      spec,
		key:       key,
		timeToAct: timeToAct,
	}, nil
}

func (l *LimiterGlobal) Cancel(r *Reservation) {
	if r == nil || r.wasCancelPrevented || r.tokenTaken == 0 {
		return
	}

	_, _, err := l.Storage.Update(r.key, r.spec.Period, r.spec.Burst, -r.tokenTaken)
	if err != nil {
		// Errors here are non-critical and non-recoverable;
		// log and continue.
		l.Logger.WithError(err).
			WithField("global", r.spec.IsGlobal).
			WithField("key", r.spec.Key()).
			Warn("failed to cancel reservation")
	}
}
