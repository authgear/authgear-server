package ratelimit

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type fakeStorage struct {
	deltas []float64
	ok     bool
}

func (s *fakeStorage) Update(ctx context.Context, key string, period time.Duration, burst int, delta float64) (bool, time.Time, error) {
	s.deltas = append(s.deltas, delta)
	return s.ok, time.Unix(0, 0).UTC(), nil
}

func TestLimiterAllowN(t *testing.T) {
	Convey("Limiter.AllowN", t, func() {
		// RateLimitGroup is left empty on purpose: doReserveN only dispatches
		// the rate_limit.blocked audit event for a named group, and dispatching
		// needs Database and EventService, which these cases do not wire.
		// Bucket arithmetic is what is under test here; the atomicity of an
		// n-token take belongs to the GCRA script and is covered by TestGCRA's
		// "burst requests" schedule.
		newSpec := func() BucketSpec {
			return BucketSpec{
				Name:    BucketName("TestBucket"),
				Enabled: true,
				Period:  1 * time.Minute,
				Burst:   10,
			}
		}

		newLimiter := func(storage *fakeStorage) *Limiter {
			return &Limiter{
				Storage: storage,
				AppID:   config.AppID("app-id"),
				Config:  &config.RateLimitsFeatureConfig{},
			}
		}

		Convey("takes n tokens in one call", func() {
			storage := &fakeStorage{ok: true}
			limiter := newLimiter(storage)

			failed, err := limiter.AllowN(context.Background(), newSpec(), 3)

			So(err, ShouldBeNil)
			So(failed, ShouldBeNil)
			So(storage.deltas, ShouldResemble, []float64{3})
		})

		Convey("AllowN with n=1 matches Allow", func() {
			storageAllowN := &fakeStorage{ok: true}
			_, err := newLimiter(storageAllowN).AllowN(context.Background(), newSpec(), 1)
			So(err, ShouldBeNil)

			storageAllow := &fakeStorage{ok: true}
			_, err = newLimiter(storageAllow).Allow(context.Background(), newSpec())
			So(err, ShouldBeNil)

			So(storageAllowN.deltas, ShouldResemble, storageAllow.deltas)
		})

		Convey("multiplies n by the group weight", func() {
			storage := &fakeStorage{ok: true}
			limiter := newLimiter(storage)

			spec := newSpec()
			spec.RateLimitGroup = RateLimitGroupAuthenticationSignup

			ctx := WithRateLimitWeights(context.Background())
			SetRateLimitWeights(ctx, Weights{RateLimitGroupAuthenticationSignup: 2})

			failed, err := limiter.AllowN(ctx, spec, 3)

			So(err, ShouldBeNil)
			So(failed, ShouldBeNil)
			So(storage.deltas, ShouldResemble, []float64{6})
		})

		Convey("returns a failed reservation when the bucket rejects the take", func() {
			storage := &fakeStorage{ok: false}
			limiter := newLimiter(storage)

			failed, err := limiter.AllowN(context.Background(), newSpec(), 3)

			So(err, ShouldBeNil)
			So(failed, ShouldNotBeNil)
			So(storage.deltas, ShouldResemble, []float64{3})
		})

		Convey("does not touch storage for a disabled bucket", func() {
			storage := &fakeStorage{ok: false}
			limiter := newLimiter(storage)

			spec := newSpec()
			spec.Enabled = false

			failed, err := limiter.AllowN(context.Background(), spec, 3)

			So(err, ShouldBeNil)
			So(failed, ShouldBeNil)
			So(storage.deltas, ShouldBeEmpty)
		})
	})
}
