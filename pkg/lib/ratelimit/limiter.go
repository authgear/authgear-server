package ratelimit

import (
	"fmt"

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

	var key string
	if spec.IsGlobal {
		key = bucketKeyGlobal(spec)
	} else {
		key = bucketKeyApp(l.AppID, spec)
	}

	ok, timeToAct, err := l.Storage.Update(key, spec.Period, spec.Burst, n)
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
		key:        key,
		ok:         ok,
		err:        err,
		tokenTaken: n,
		timeToAct:  &timeToAct,
	}
}

func (l *Limiter) Cancel(r *Reservation) {
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

func bucketKeyGlobal(spec BucketSpec) string {
	return fmt.Sprintf("rate-limit:%s", spec.Key())
}

func bucketKeyApp(appID config.AppID, spec BucketSpec) string {
	return fmt.Sprintf("app:%s:rate-limit:%s", appID, spec.Key())
}
