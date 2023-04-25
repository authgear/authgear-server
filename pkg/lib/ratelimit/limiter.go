package ratelimit

import (
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
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
	Clock   clock.Clock
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

	now := l.Clock.NowUTC()

	pass := false
	tokenTaken := 0
	var timeToAct *time.Time
	err := l.Storage.WithConn(func(conn StorageConn) error {
		// Check if we should refill the bucket.
		resetTime, err := conn.GetResetTime(spec, now)
		if err != nil {
			return err
		}
		if !now.Before(resetTime) {
			// Refill bucket to full.
			err = conn.Reset(spec, now)
			if err != nil {
				return err
			}
			resetTime = now
		}

		// Try to take requested tokens.
		tokens, err := conn.TakeToken(spec, now, n)
		if err != nil {
			return err
		}
		tokenTaken = n

		pass = tokens >= 0
		timeToAct = &resetTime
		l.Logger.
			WithField("global", spec.IsGlobal).
			WithField("key", spec.Key()).
			WithField("tokens", tokens).
			WithField("pass", pass).
			Debug("check rate limit")
		return nil
	})

	return &Reservation{
		spec:       spec,
		ok:         pass,
		err:        err,
		tokenTaken: tokenTaken,
		timeToAct:  timeToAct,
	}
}

func (l *Limiter) Cancel(r *Reservation) {
	if r == nil || r.isConsumed || r.tokenTaken == 0 {
		return
	}

	now := l.Clock.NowUTC()

	err := l.Storage.WithConn(func(conn StorageConn) error {
		// Check if we should refill the bucket.
		resetTime, err := conn.GetResetTime(r.spec, now)
		if err != nil {
			return err
		}
		if !now.Before(resetTime) {
			// Refill bucket to full; no need further restore tokens
			err = conn.Reset(r.spec, now)
			return err
		}

		// Try to put all taken tokens.
		_, err = conn.TakeToken(r.spec, now, -r.tokenTaken)
		if err != nil {
			return err
		}

		r.tokenTaken = 0
		return nil
	})

	if err != nil {
		// Errors here are non-critical and non-recoverable;
		// log and continue.
		l.Logger.WithError(err).
			WithField("global", r.spec.IsGlobal).
			WithField("key", r.spec.Key()).
			Warn("failed to cancel reservation")
	}
}
