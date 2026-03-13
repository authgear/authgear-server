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

func newMetricsStore(t *testing.T, clk clock.Clock) (*MetricsStore, *miniredis.Miniredis) {
	mr := miniredis.RunT(t)

	pool := redis.NewPool()
	rh := redis.NewHandle(pool, redis.ConnectionOptions{
		RedisURL:              "redis://" + mr.Addr(),
		MaxOpenConnection:     func(i int) *int { return &i }(10),
		MaxIdleConnection:     func(i int) *int { return &i }(5),
		IdleConnectionTimeout: func(d config.DurationSeconds) *config.DurationSeconds { return &d }(300),
		MaxConnectionLifetime: func(d config.DurationSeconds) *config.DurationSeconds { return &d }(900),
	})

	store := &MetricsStore{
		Redis: &appredis.Handle{Handle: rh},
		AppID: config.AppID("test-app"),
		Clock: clk,
	}
	return store, mr
}

func TestMetricsStoreThresholdCacheKey(t *testing.T) {
	Convey("thresholdCacheKey", t, func() {
		s := &MetricsStore{AppID: "myapp"}

		So(s.thresholdCacheKey("phone_country:SG", "24h"),
			ShouldEqual,
			"app:myapp:fraud_protection:threshold_cache:sms_otp_verified:24h:phone_country:SG")

		So(s.thresholdCacheKey("ip:1.2.3.4", "24h"),
			ShouldEqual,
			"app:myapp:fraud_protection:threshold_cache:sms_otp_verified:24h:ip:1.2.3.4")

		So(s.thresholdCacheKey("phone_country:MY", "14d_max"),
			ShouldEqual,
			"app:myapp:fraud_protection:threshold_cache:sms_otp_verified:14d_max:phone_country:MY")

		So(s.thresholdCacheKey("phone_country:SG", "1h"),
			ShouldEqual,
			"app:myapp:fraud_protection:threshold_cache:sms_otp_verified:1h:phone_country:SG")
	})
}

func TestMetricsStoreCacheRoundTrip(t *testing.T) {
	Convey("getCachedCount and setCachedCount", t, func() {
		ctx := context.Background()
		now := time.Unix(testEpoch, 0).UTC()
		clk := clock.NewMockClockAtTime(now)

		store, _ := newMetricsStore(t, clk)

		cacheKey := store.thresholdCacheKey("phone_country:SG", "24h")

		Convey("returns redis.Nil when key does not exist", func() {
			_, err := store.getCachedCount(ctx, cacheKey)
			So(err, ShouldNotBeNil)
		})

		Convey("round-trips a stored value", func() {
			err := store.setCachedCount(ctx, cacheKey, 42)
			So(err, ShouldBeNil)

			count, err := store.getCachedCount(ctx, cacheKey)
			So(err, ShouldBeNil)
			So(count, ShouldEqual, 42)
		})

		Convey("round-trips zero correctly", func() {
			err := store.setCachedCount(ctx, cacheKey, 0)
			So(err, ShouldBeNil)

			count, err := store.getCachedCount(ctx, cacheKey)
			So(err, ShouldBeNil)
			So(count, ShouldEqual, 0)
		})
	})
}

func TestMetricsStoreGetVerifiedCacheHit(t *testing.T) {
	Convey("Get methods return cached value without hitting DB", t, func() {
		ctx := context.Background()
		now := time.Unix(testEpoch, 0).UTC()
		clk := clock.NewMockClockAtTime(now)

		// DB fields are nil — if code reaches the PostgreSQL path it will panic.
		// A cache hit must occur so that the DB is never reached.
		store, _ := newMetricsStore(t, clk)

		Convey("GetVerifiedByCountry24h returns cached count", func() {
			cacheKey := store.thresholdCacheKey("phone_country:SG", "24h")
			_ = store.setCachedCount(ctx, cacheKey, 55)

			count, err := store.GetVerifiedByCountry24h(ctx, "SG")
			So(err, ShouldBeNil)
			So(count, ShouldEqual, 55)
		})

		Convey("GetVerifiedByCountry1h returns cached count", func() {
			cacheKey := store.thresholdCacheKey("phone_country:SG", "1h")
			_ = store.setCachedCount(ctx, cacheKey, 7)

			count, err := store.GetVerifiedByCountry1h(ctx, "SG")
			So(err, ShouldBeNil)
			So(count, ShouldEqual, 7)
		})

		Convey("GetVerifiedByIP24h returns cached count", func() {
			cacheKey := store.thresholdCacheKey("ip:1.2.3.4", "24h")
			_ = store.setCachedCount(ctx, cacheKey, 3)

			count, err := store.GetVerifiedByIP24h(ctx, "1.2.3.4")
			So(err, ShouldBeNil)
			So(count, ShouldEqual, 3)
		})

		Convey("GetVerifiedByCountryPast14DaysRollingMax returns cached count", func() {
			cacheKey := store.thresholdCacheKey("phone_country:SG", "14d_max")
			_ = store.setCachedCount(ctx, cacheKey, 100)

			count, err := store.GetVerifiedByCountryPast14DaysRollingMax(ctx, "SG")
			So(err, ShouldBeNil)
			So(count, ShouldEqual, 100)
		})
	})
}
