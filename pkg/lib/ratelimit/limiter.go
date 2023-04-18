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

func (l *Limiter) isDisabled() bool {
	return l.Config.Disabled
}

func (l *Limiter) TakeToken(bucket Bucket) error {
	if l.isDisabled() {
		return nil
	}

	return l.Storage.WithConn(func(conn StorageConn) error {
		now := l.Clock.NowUTC()

		// Check if we should refill the bucket.
		resetTime, err := conn.GetResetTime(bucket, now)
		if err != nil {
			return err
		}
		if !now.Before(resetTime) {
			// Refill bucket to full.
			err = conn.Reset(bucket, now)
			if err != nil {
				return err
			}
		}

		// Try to take one token.
		tokens, err := conn.TakeToken(bucket, now, 1)
		if err != nil {
			return err
		}

		pass := tokens >= 0
		l.Logger.
			WithField("key", bucket.Key).
			WithField("tokens", tokens).
			WithField("pass", pass).
			Debug("check rate limit")

		if !pass {
			// Exhausted tokens, rate limit the request.
			return bucket.BucketError()
		}

		return nil
	})
}

// CheckToken return resetDuration and pass based on the given bucket
func (l *Limiter) CheckToken(bucket Bucket) (pass bool, resetDuration time.Duration, err error) {
	if l.isDisabled() {
		return true, time.Duration(0), nil
	}

	err = l.Storage.WithConn(func(conn StorageConn) error {
		now := l.Clock.NowUTC()

		resetTime, err := conn.GetResetTime(bucket, now)
		if err != nil {
			return err
		}
		if !now.Before(resetTime) {
			// Exceed the reset time, bucket will be reset
			pass = true
			return nil
		}

		// Check the token
		tokens, err := conn.CheckToken(bucket)
		if err != nil {
			return err
		}

		resetDuration = resetTime.Sub(now)
		// We need at least 1 token to consume next time.
		pass = tokens >= 1
		return nil
	})

	return
}

// RequireToken requires the bucket to have at least one token
func (l *Limiter) RequireToken(bucket Bucket) error {
	pass, _, err := l.CheckToken(bucket)
	if err != nil {
		return err
	}

	if !pass {
		return bucket.BucketError()
	}

	return nil
}

func (l *Limiter) ClearBucket(bucket Bucket) error {
	return l.Storage.WithConn(func(conn StorageConn) error {
		now := l.Clock.NowUTC()
		return conn.Reset(bucket, now)
	})
}

func (l *Limiter) Allow(spec BucketSpec) error {
	if !spec.Enabled {
		return nil
	}
	return l.TakeToken(spec.bucket())
}

func (l *Limiter) Reserve(spec BucketSpec) *Reservation {
	return l.ReserveN(spec, 1)
}

func (l *Limiter) ReserveN(spec BucketSpec, n int) *Reservation {
	if l.isDisabled() || !spec.Enabled {
		return &Reservation{spec: spec, ok: true}
	}

	bucket := spec.bucket()
	now := l.Clock.NowUTC()

	pass := false
	tokenTaken := 0
	var timeToAct *time.Time
	err := l.Storage.WithConn(func(conn StorageConn) error {
		// Check if we should refill the bucket.
		resetTime, err := conn.GetResetTime(bucket, now)
		if err != nil {
			return err
		}
		if !now.Before(resetTime) {
			// Refill bucket to full.
			err = conn.Reset(bucket, now)
			if err != nil {
				return err
			}
			resetTime = now
		}

		// Try to take requested tokens.
		tokens, err := conn.TakeToken(bucket, now, n)
		if err != nil {
			return err
		}
		tokenTaken = n

		pass = tokens >= 0
		timeToAct = &resetTime
		l.Logger.
			WithField("global", bucket.IsGlobal).
			WithField("key", bucket.Key).
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

func (l *Limiter) Cancel(r *Reservation) error {
	if r == nil || r.isConsumed || r.tokenTaken == 0 {
		return nil
	}

	bucket := r.spec.bucket()
	now := l.Clock.NowUTC()

	return l.Storage.WithConn(func(conn StorageConn) error {
		// Check if we should refill the bucket.
		resetTime, err := conn.GetResetTime(bucket, now)
		if err != nil {
			return err
		}
		if !now.Before(resetTime) {
			// Refill bucket to full; no need further restore tokens
			err = conn.Reset(bucket, now)
			return err
		}

		// Try to put all taken tokens.
		_, err = conn.TakeToken(bucket, now, -r.tokenTaken)
		if err != nil {
			return err
		}

		r.tokenTaken = 0
		return nil
	})
}
