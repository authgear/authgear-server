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

func (l *Limiter) Allow(spec BucketSpec) (*FailedReservation, error) {
	_, failedReservation, err := l.Reserve(spec)
	return failedReservation, err
}

func (l *Limiter) Reserve(spec BucketSpec) (*Reservation, *FailedReservation, error) {
	return l.ReserveN(spec, 1)
}

func (l *Limiter) ReserveN(spec BucketSpec, n int) (*Reservation, *FailedReservation, error) {
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
		}, nil, nil
	}

	ok, timeToAct, err := l.Storage.Update(key, spec.Period, spec.Burst, n)
	if err != nil {
		return nil, nil, err
	}

	l.Logger.
		WithField("global", spec.IsGlobal).
		WithField("key", key).
		WithField("ok", ok).
		WithField("timeToAct", timeToAct).
		Debug("check rate limit")

	if ok {
		return &Reservation{
			key:        key,
			spec:       spec,
			tokenTaken: n,
		}, nil, nil
	}

	return nil, &FailedReservation{
		key:       key,
		spec:      spec,
		timeToAct: timeToAct,
	}, nil
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
