package ratelimit_test

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
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

func (s *storageMemory) CheckToken(bucket ratelimit.Bucket) (int, error) {
	data, ok := s.items[bucket.Key]
	if ok {
		return bucket.Size - data.tokenTaken, nil
	}
	return bucket.Size, nil
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
		expectedErr := b.BucketError()

		Convey("TakeToken", func() {
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
			So(err, ShouldBeError, expectedErr)
			err = limiter.TakeToken(b)
			So(err, ShouldBeError, expectedErr)
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
			So(err, ShouldBeError, expectedErr)
			c.AdvanceSeconds(1)

			err = limiter.TakeToken(b)
			So(err, ShouldBeError, expectedErr)
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
			So(err, ShouldBeError, expectedErr)
			c.AdvanceSeconds(1)

			err = limiter.TakeToken(b)
			So(err, ShouldBeError, expectedErr)
			c.AdvanceSeconds(1)
		})

		Convey("CheckToken", func() {
			var err error
			var resetDuration time.Duration
			var pass bool

			_ = limiter.TakeToken(b)
			_ = limiter.TakeToken(b)
			pass, resetDuration, err = limiter.CheckToken(b)
			So(resetDuration, ShouldEqual, 5*time.Second)
			So(pass, ShouldBeTrue)
			So(err, ShouldBeNil)
			c.AdvanceSeconds(1)

			err = limiter.TakeToken(b)
			So(err, ShouldBeNil)
			pass, resetDuration, err = limiter.CheckToken(b)
			So(resetDuration, ShouldEqual, 4*time.Second)
			So(pass, ShouldBeFalse)
			So(err, ShouldBeNil)

			c.AdvanceSeconds(4)

			// Reset

			pass, _, err = limiter.CheckToken(b)
			So(pass, ShouldBeTrue)
			So(err, ShouldBeNil)

			err = limiter.TakeToken(b)
			So(err, ShouldBeNil)
			pass, resetDuration, err = limiter.CheckToken(b)
			So(resetDuration, ShouldEqual, 5*time.Second)
			So(pass, ShouldBeTrue)
			So(err, ShouldBeNil)
			c.AdvanceSeconds(1)

			err = limiter.TakeToken(b)
			So(err, ShouldBeNil)
			pass, resetDuration, err = limiter.CheckToken(b)
			So(resetDuration, ShouldEqual, 4*time.Second)
			So(pass, ShouldBeTrue)
			So(err, ShouldBeNil)
			c.AdvanceSeconds(1)

			err = limiter.TakeToken(b)
			So(err, ShouldBeNil)
			pass, resetDuration, err = limiter.CheckToken(b)
			So(resetDuration, ShouldEqual, 3*time.Second)
			So(pass, ShouldBeFalse)
			So(err, ShouldBeNil)
			c.AdvanceSeconds(3)

			// Reset

			pass, _, err = limiter.CheckToken(b)
			So(pass, ShouldBeTrue)
			So(err, ShouldBeNil)

			err = limiter.TakeToken(b)
			So(err, ShouldBeNil)
			c.AdvanceSeconds(5)

			// Reset

			pass, _, err = limiter.CheckToken(b)
			So(pass, ShouldBeTrue)
			So(err, ShouldBeNil)

			err = limiter.TakeToken(b)
			So(err, ShouldBeNil)
			pass, resetDuration, err = limiter.CheckToken(b)
			So(resetDuration, ShouldEqual, 5*time.Second)
			So(pass, ShouldBeTrue)
			So(err, ShouldBeNil)

			err = limiter.TakeToken(b)
			So(err, ShouldBeNil)
			pass, resetDuration, err = limiter.CheckToken(b)
			So(resetDuration, ShouldEqual, 5*time.Second)
			So(pass, ShouldBeTrue)
			So(err, ShouldBeNil)

			c.AdvanceSeconds(1)

			err = limiter.TakeToken(b)
			So(err, ShouldBeNil)
			pass, resetDuration, err = limiter.CheckToken(b)
			So(resetDuration, ShouldEqual, 4*time.Second)
			So(pass, ShouldBeFalse)
			So(err, ShouldBeNil)

			c.AdvanceSeconds(1)

			pass, resetDuration, err = limiter.CheckToken(b)
			So(resetDuration, ShouldEqual, 3*time.Second)
			So(pass, ShouldBeFalse)
			So(err, ShouldBeNil)
		})

		Convey("Disabled", func() {
			disabled := true
			limiter.Config = &config.RateLimitFeatureConfig{
				Disabled: &disabled,
			}

			for i := 0; i < 2*b.Size; i++ {
				pass, _, err := limiter.CheckToken(b)
				So(pass, ShouldBeTrue)
				So(err, ShouldBeNil)

				err = limiter.TakeToken(b)
				So(err, ShouldBeNil)
			}
		})
	})
}
