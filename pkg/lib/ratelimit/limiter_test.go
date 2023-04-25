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

func (s *storageMemory) TakeToken(spec ratelimit.BucketSpec, now time.Time, delta int) (int, error) {
	data, ok := s.items[spec.Key()]
	if ok {
		data.tokenTaken += delta
		return spec.Burst - data.tokenTaken, nil
	}

	data = &storageMemoryItem{
		tokenTaken: 1,
		resetTime:  now.Add(spec.Period),
	}
	s.items[spec.Key()] = data
	return spec.Burst - data.tokenTaken, nil
}

func (s *storageMemory) GetResetTime(spec ratelimit.BucketSpec, now time.Time) (time.Time, error) {
	data, ok := s.items[spec.Key()]
	if ok {
		return data.resetTime, nil
	}
	return now, nil
}

func (s *storageMemory) Reset(spec ratelimit.BucketSpec, now time.Time) error {
	data := &storageMemoryItem{
		tokenTaken: 0,
		resetTime:  now.Add(spec.Period),
	}
	s.items[spec.Key()] = data
	return nil
}

func TestLimiter(t *testing.T) {
	Convey("Limiter", t, func() {
		b := ratelimit.BucketSpec{Name: "BucketA", Enabled: true, Burst: 3, Period: 5 * time.Second}
		c := clock.NewMockClock()
		limiter := &ratelimit.Limiter{
			Logger:  ratelimit.Logger{log.Null},
			Storage: &storageMemory{items: make(map[string]*storageMemoryItem)},
			Clock:   c,
			Config:  &config.RateLimitsFeatureConfig{},
		}
		expectedErr := ratelimit.ErrRateLimited("BucketA")

		Convey("TakeToken", func() {
			var err error

			err = limiter.Allow(b)
			So(err, ShouldBeNil)
			err = limiter.Allow(b)
			So(err, ShouldBeNil)
			c.AdvanceSeconds(1)

			err = limiter.Allow(b)
			So(err, ShouldBeNil)
			c.AdvanceSeconds(1)

			c.AdvanceSeconds(1)

			err = limiter.Allow(b)
			So(err, ShouldBeError, expectedErr)
			err = limiter.Allow(b)
			So(err, ShouldBeError, expectedErr)
			c.AdvanceSeconds(1)

			c.AdvanceSeconds(1)

			// Reset

			err = limiter.Allow(b)
			So(err, ShouldBeNil)
			c.AdvanceSeconds(1)

			err = limiter.Allow(b)
			So(err, ShouldBeNil)
			c.AdvanceSeconds(1)

			err = limiter.Allow(b)
			So(err, ShouldBeNil)
			c.AdvanceSeconds(1)

			err = limiter.Allow(b)
			So(err, ShouldBeError, expectedErr)
			c.AdvanceSeconds(1)

			err = limiter.Allow(b)
			So(err, ShouldBeError, expectedErr)
			c.AdvanceSeconds(1)

			// Reset

			err = limiter.Allow(b)
			So(err, ShouldBeNil)
			c.AdvanceSeconds(1)

			c.AdvanceSeconds(1)

			c.AdvanceSeconds(1)

			c.AdvanceSeconds(1)

			c.AdvanceSeconds(1)

			// Reset

			err = limiter.Allow(b)
			So(err, ShouldBeNil)
			err = limiter.Allow(b)
			So(err, ShouldBeNil)
			c.AdvanceSeconds(1)

			err = limiter.Allow(b)
			So(err, ShouldBeNil)
			c.AdvanceSeconds(1)

			err = limiter.Allow(b)
			So(err, ShouldBeError, expectedErr)
			c.AdvanceSeconds(1)

			err = limiter.Allow(b)
			So(err, ShouldBeError, expectedErr)
			c.AdvanceSeconds(1)
		})

		Convey("Disabled", func() {
			limiter.Config = &config.RateLimitsFeatureConfig{
				Disabled: true,
			}

			for i := 0; i < 2*b.Burst; i++ {
				err := limiter.Allow(b)
				So(err, ShouldBeNil)
			}
		})
	})
}
