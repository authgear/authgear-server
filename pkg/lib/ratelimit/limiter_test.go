package ratelimit_test

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type storageMemoryItem struct {
	tokenTaken int
	resetTime  time.Time
}

type storageMemory struct {
	items map[string]*storageMemoryItem
}

func (s *storageMemory) WithConn(f func(ratelimit.StorageConn) error) error {
	return f(s)
}

func (s *storageMemory) TakeToken(bucket ratelimit.Bucket, now time.Time) (int, error) {
	data, ok := s.items[bucket.Key]
	if ok {
		data.tokenTaken++
		return bucket.Size - data.tokenTaken, nil
	}

	data = &storageMemoryItem{
		tokenTaken: 1,
		resetTime:  now.Add(bucket.ResetPeriod),
	}
	s.items[bucket.Key] = data
	return bucket.Size - data.tokenTaken, nil
}

func (s *storageMemory) GetResetTime(bucket ratelimit.Bucket, now time.Time) (time.Time, error) {
	data, ok := s.items[bucket.Key]
	if ok {
		return data.resetTime, nil
	}
	return now, nil
}

func (s *storageMemory) Reset(bucket ratelimit.Bucket, now time.Time) error {
	data := &storageMemoryItem{
		tokenTaken: 0,
		resetTime:  now.Add(bucket.ResetPeriod),
	}
	s.items[bucket.Key] = data
	return nil
}

func TestLimiter(t *testing.T) {
	Convey("Limiter", t, func() {
		b := ratelimit.Bucket{Key: "bucket", Size: 3, ResetPeriod: 5 * time.Second}
		c := clock.NewMockClock()
		limiter := &ratelimit.Limiter{
			Logger:  ratelimit.Logger{log.Null},
			Storage: &storageMemory{items: make(map[string]*storageMemoryItem)},
			Clock:   c,
		}

		var err error

		err = limiter.TakeToken(b)
		So(err, ShouldBeNil)
		err = limiter.TakeToken(b)
		So(err, ShouldBeNil)
		c.AdvanceSeconds(1)

		err = limiter.TakeToken(b)
		So(err, ShouldBeNil)
		c.AdvanceSeconds(1)

		c.AdvanceSeconds(1)

		err = limiter.TakeToken(b)
		So(err, ShouldBeError, ratelimit.ErrTooManyRequests)
		err = limiter.TakeToken(b)
		So(err, ShouldBeError, ratelimit.ErrTooManyRequests)
		c.AdvanceSeconds(1)

		c.AdvanceSeconds(1)

		// Reset

		err = limiter.TakeToken(b)
		So(err, ShouldBeNil)
		c.AdvanceSeconds(1)

		err = limiter.TakeToken(b)
		So(err, ShouldBeNil)
		c.AdvanceSeconds(1)

		err = limiter.TakeToken(b)
		So(err, ShouldBeNil)
		c.AdvanceSeconds(1)

		err = limiter.TakeToken(b)
		So(err, ShouldBeError, ratelimit.ErrTooManyRequests)
		c.AdvanceSeconds(1)

		err = limiter.TakeToken(b)
		So(err, ShouldBeError, ratelimit.ErrTooManyRequests)
		c.AdvanceSeconds(1)

		// Reset

		err = limiter.TakeToken(b)
		So(err, ShouldBeNil)
		c.AdvanceSeconds(1)

		c.AdvanceSeconds(1)

		c.AdvanceSeconds(1)

		c.AdvanceSeconds(1)

		c.AdvanceSeconds(1)

		// Reset

		err = limiter.TakeToken(b)
		So(err, ShouldBeNil)
		err = limiter.TakeToken(b)
		So(err, ShouldBeNil)
		c.AdvanceSeconds(1)

		err = limiter.TakeToken(b)
		So(err, ShouldBeNil)
		c.AdvanceSeconds(1)

		err = limiter.TakeToken(b)
		So(err, ShouldBeError, ratelimit.ErrTooManyRequests)
		c.AdvanceSeconds(1)

		err = limiter.TakeToken(b)
		So(err, ShouldBeError, ratelimit.ErrTooManyRequests)
		c.AdvanceSeconds(1)
	})
}
