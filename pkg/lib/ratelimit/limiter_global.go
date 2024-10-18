package ratelimit

import "errors"

type LimiterGlobal struct {
	Logger  Logger
	Storage Storage
}

func (l *LimiterGlobal) Allow(spec BucketSpec) error {
	r := l.Reserve(spec)
	return r.Error()
}

func (l *LimiterGlobal) Reserve(spec BucketSpec) *Reservation {
	return l.ReserveN(spec, 1)
}

func (l *LimiterGlobal) ReserveN(spec BucketSpec, n int) *Reservation {
	if !spec.IsGlobal {
		return &Reservation{
			spec: spec,
			ok:   false,
			err:  errors.New("ratelimit: must be global limit"),
		}
	}

	if !spec.Enabled {
		return &Reservation{spec: spec, ok: true}
	}

	key := bucketKeyGlobal(spec)
	ok, timeToAct, err := l.Storage.Update(key, spec.Period, spec.Burst, n)
	if err != nil {
		return &Reservation{
			spec: spec,
			ok:   false,
			err:  err,
		}
	}

	l.Logger.
		WithField("key", spec.Key()).
		WithField("ok", ok).
		WithField("timeToAct", timeToAct).
		Debug("check global rate limit")

	return &Reservation{
		spec:       spec,
		key:        key,
		ok:         ok,
		err:        err,
		tokenTaken: n,
		timeToAct:  &timeToAct,
	}
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
