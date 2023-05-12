package ratelimit

import (
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
	Config  *config.RateLimitsFeatureConfig
}

func (l *Limiter) Allow(spec BucketSpec) error {
	r := l.Reserve(spec)
	return r.Error()
}

func (l *Limiter) Reserve(spec BucketSpec) *Reservation {
	return l.ReserveN(spec, 1)
}

func (l *Limiter) ReserveN(spec BucketSpec, n int) *Reservation {
	if l.Config.Disabled || !spec.Enabled {
		return &Reservation{spec: spec, ok: true}
	}

	ok, timeToAct, err := l.Storage.Update(spec, n)
	if err != nil {
		return &Reservation{
			spec: spec,
			ok:   false,
			err:  err,
		}
	}

	l.Logger.
		WithField("global", spec.IsGlobal).
		WithField("key", spec.Key()).
		WithField("ok", ok).
		WithField("timeToAct", timeToAct).
		Debug("check rate limit")

	return &Reservation{
		spec:       spec,
		ok:         ok,
		err:        err,
		tokenTaken: n,
		timeToAct:  &timeToAct,
	}
}

func (l *Limiter) Cancel(r *Reservation) {
	if r == nil || r.isConsumed || r.tokenTaken == 0 {
		return
	}

	_, _, err := l.Storage.Update(r.spec, -r.tokenTaken)
	if err != nil {
		// Errors here are non-critical and non-recoverable;
		// log and continue.
		l.Logger.WithError(err).
			WithField("global", r.spec.IsGlobal).
			WithField("key", r.spec.Key()).
			Warn("failed to cancel reservation")
	}
}
