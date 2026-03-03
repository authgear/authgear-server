package fraudprotection

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

const testEpoch = int64(1_700_000_000) // arbitrary fixed timestamp

func newLeakyBucketStore(t *testing.T, clk clock.Clock) (*LeakyBucketStore, *miniredis.Miniredis) {
	mr := miniredis.RunT(t)

	pool := redis.NewPool()
	rh := redis.NewHandle(pool, redis.ConnectionOptions{
		RedisURL:              "redis://" + mr.Addr(),
		MaxOpenConnection:     func(i int) *int { return &i }(10),
		MaxIdleConnection:     func(i int) *int { return &i }(5),
		IdleConnectionTimeout: func(d config.DurationSeconds) *config.DurationSeconds { return &d }(300),
		MaxConnectionLifetime: func(d config.DurationSeconds) *config.DurationSeconds { return &d }(900),
	})

	store := &LeakyBucketStore{
		Redis: &appredis.Handle{Handle: rh},
		AppID: config.AppID("test-app"),
		Clock: clk,
	}
	return store, mr
}

func TestLeakyBucketStoreKeys(t *testing.T) {
	Convey("Key helpers", t, func() {
		s := &LeakyBucketStore{AppID: "myapp"}

		Convey("bucketKey", func() {
			So(s.bucketKey(bucketWindowHourly, bucketDimensionCountry, "SG"), ShouldEqual, "app:myapp:fraud_protection:leaky_bucket:3600:country:SG")
			So(s.bucketKey(bucketWindowDaily, bucketDimensionIP, "1.2.3.4"), ShouldEqual, "app:myapp:fraud_protection:leaky_bucket:86400:ip:1.2.3.4")
		})

		Convey("ipCountriesKey", func() {
			So(s.ipCountriesKey("1.2.3.4"), ShouldEqual, "app:myapp:fraud_protection:ip_countries:1.2.3.4")
		})
	})
}

func TestLeakyBucketStoreRecordSMSOTPSent(t *testing.T) {
	Convey("RecordSMSOTPSent", t, func() {
		ctx := context.Background()
		now := time.Unix(testEpoch, 0).UTC()
		clk := clock.NewMockClockAtTime(now)

		store, mr := newLeakyBucketStore(t, clk)
		mr.SetTime(now)

		thresholds := LeakyBucketThresholds{
			CountryHourly: 3,
			CountryDaily:  20,
			IPHourly:      5,
			IPDaily:       10,
		}

		Convey("no bucket is triggered when level is below threshold", func() {
			triggered, err := store.RecordSMSOTPSent(ctx, "1.2.3.4", "SG", thresholds)
			So(err, ShouldBeNil)
			So(triggered.CountryHourly, ShouldBeFalse)
			So(triggered.CountryDaily, ShouldBeFalse)
			So(triggered.IPHourly, ShouldBeFalse)
			So(triggered.IPDaily, ShouldBeFalse)
			So(triggered.IPCountriesDaily, ShouldBeFalse)
		})

		Convey("IPHourly triggers after exceeding IP hourly threshold", func() {
			// ipHourly threshold = 5; send 6 times without time advancing
			for i := 0; i < 5; i++ {
				_, err := store.RecordSMSOTPSent(ctx, "1.2.3.4", "SG", thresholds)
				So(err, ShouldBeNil)
			}
			triggered, err := store.RecordSMSOTPSent(ctx, "1.2.3.4", "SG", thresholds)
			So(err, ShouldBeNil)
			So(triggered.IPHourly, ShouldBeTrue)
		})

		Convey("CountryHourly triggers after exceeding country hourly threshold", func() {
			// countryHourly threshold = 3; use different IPs to avoid IP bucket interference
			for i := 0; i < 3; i++ {
				_, err := store.RecordSMSOTPSent(ctx, "10.0.0.1", "SG", thresholds)
				So(err, ShouldBeNil)
			}
			triggered, err := store.RecordSMSOTPSent(ctx, "10.0.0.2", "SG", thresholds)
			So(err, ShouldBeNil)
			So(triggered.CountryHourly, ShouldBeTrue)
		})

		Convey("IPCountriesDaily triggers when IP contacts more than 3 countries in 24h", func() {
			countries := []string{"SG", "MY", "TH", "US"}
			for i, country := range countries {
				triggered, err := store.RecordSMSOTPSent(ctx, "1.2.3.4", country, thresholds)
				So(err, ShouldBeNil)
				if i < 3 {
					So(triggered.IPCountriesDaily, ShouldBeFalse)
				} else {
					So(triggered.IPCountriesDaily, ShouldBeTrue)
				}
			}
		})

		Convey("level drains over time and no longer triggers after sufficient drain", func() {
			// Fill country hourly bucket (threshold=3) past the threshold.
			for i := 0; i < 4; i++ {
				_, _ = store.RecordSMSOTPSent(ctx, "1.2.3.4", "SG", thresholds)
			}

			// Advance time by 1 full hour — the bucket should fully drain
			// (drain_rate = 3/3600, so after 3600s the level drops by 3 from wherever it was).
			later := now.Add(2 * time.Hour)
			mr.SetTime(later)
			mr.FastForward(2 * time.Hour)
			clk.Time = later

			triggered, err := store.RecordSMSOTPSent(ctx, "1.2.3.4", "SG", thresholds)
			So(err, ShouldBeNil)
			So(triggered.CountryHourly, ShouldBeFalse)
		})
	})
}

func TestLeakyBucketStoreRecordSMSOTPVerified(t *testing.T) {
	Convey("RecordSMSOTPVerified", t, func() {
		ctx := context.Background()
		now := time.Unix(testEpoch, 0).UTC()
		clk := clock.NewMockClockAtTime(now)

		store, mr := newLeakyBucketStore(t, clk)
		mr.SetTime(now)

		thresholds := LeakyBucketThresholds{
			CountryHourly: 3,
			CountryDaily:  20,
			IPHourly:      5,
			IPDaily:       10,
		}

		Convey("draining after fills brings bucket back below threshold", func() {
			// Fill IP hourly bucket past threshold (threshold=5, send 6).
			for i := 0; i < 6; i++ {
				_, _ = store.RecordSMSOTPSent(ctx, "1.2.3.4", "SG", thresholds)
			}

			// Drain by 6 — level should drop to ~0.
			err := store.RecordSMSOTPVerified(ctx, "1.2.3.4", "SG", thresholds, 6)
			So(err, ShouldBeNil)

			// Next send should not trigger.
			triggered, err := store.RecordSMSOTPSent(ctx, "1.2.3.4", "SG", thresholds)
			So(err, ShouldBeNil)
			So(triggered.IPHourly, ShouldBeFalse)
		})

		Convey("draining an empty bucket does not error", func() {
			err := store.RecordSMSOTPVerified(ctx, "1.2.3.4", "SG", thresholds, 3)
			So(err, ShouldBeNil)
		})
	})
}
