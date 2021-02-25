package ratelimit

import (
	"time"

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
}

func (l *Limiter) TakeToken(bucket Bucket) error {
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
		tokens, err := conn.TakeToken(bucket, now)
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
			return ErrTooManyRequests
		}

		return nil
	})
}

// CheckToken return resetDuration and pass based on the given bucket
func (l *Limiter) CheckToken(bucket Bucket) (pass bool, resetDuration time.Duration, err error) {
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
