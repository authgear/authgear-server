package usage

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

func TestComputeResetTime(t *testing.T) {
	Convey("compute reset time of quota", t, func() {
		test := func(now string, day string, month string) {
			t, _ := time.Parse(time.RFC3339, now)
			So(
				ComputeResetTime(t, model.UsageLimitPeriodDay).Format(time.RFC3339),
				ShouldEqual,
				day,
			)
			So(
				ComputeResetTime(t, model.UsageLimitPeriodMonth).Format(time.RFC3339),
				ShouldEqual,
				month,
			)
		}
		test("2009-11-10T15:00:00Z", "2009-11-11T00:00:00Z", "2009-12-01T00:00:00Z")
		test("2009-02-28T23:00:00Z", "2009-03-01T00:00:00Z", "2009-03-01T00:00:00Z")
	})
}

const testKey = "rate-limit"
const epoch = 1257894000000

func TestLimitReserve(t *testing.T) {
	s := miniredis.RunT(t)

	Convey("reserve tokens", t, func() {
		ctx := context.Background()
		s.FlushAll()

		cli := goredis.NewClient(&goredis.Options{Addr: s.Addr()})
		conn := cli.Conn()

		now := time.UnixMilli(epoch).UTC()
		s.SetTime(now)

		forward := func(period time.Duration) {
			newNow := now.Add(period)
			s.SetTime(newNow)
			s.FastForward(newNow.Sub(now))
			now = newNow
		}

		quota := 10
		resetTime := time.UnixMilli(epoch).UTC().Add(12 * time.Hour)

		pass, before, after, err := runReserveScript(ctx, conn, testKey, 1, resetTime, quota)
		So(err, ShouldBeNil)
		So(pass, ShouldBeTrue)
		So(before, ShouldEqual, 0)
		So(after, ShouldEqual, 1)

		pass, before, after, err = runReserveScript(ctx, conn, testKey, 8, resetTime, quota)
		So(err, ShouldBeNil)
		So(pass, ShouldBeTrue)
		So(before, ShouldEqual, 1)
		So(after, ShouldEqual, 9)

		// Should not deduct tokens if the reservation attempt exceed it.
		pass, before, after, err = runReserveScript(ctx, conn, testKey, 2, resetTime, quota)
		So(err, ShouldBeNil)
		So(pass, ShouldBeFalse)
		So(before, ShouldEqual, 9)
		So(after, ShouldEqual, 9)

		pass, before, after, err = runReserveScript(ctx, conn, testKey, 1, resetTime, quota)
		So(err, ShouldBeNil)
		So(pass, ShouldBeTrue)
		So(before, ShouldEqual, 9)
		So(after, ShouldEqual, 10)

		pass, before, after, err = runReserveScript(ctx, conn, testKey, 1, resetTime, quota)
		So(err, ShouldBeNil)
		So(pass, ShouldBeFalse)
		So(before, ShouldEqual, 10)
		So(after, ShouldEqual, 10)

		forward(6 * time.Hour)

		pass, before, after, err = runReserveScript(ctx, conn, testKey, 1, resetTime, quota)
		So(err, ShouldBeNil)
		So(pass, ShouldBeFalse)
		So(before, ShouldEqual, 10)
		So(after, ShouldEqual, 10)

		forward(6 * time.Hour)

		pass, before, after, err = runReserveScript(ctx, conn, testKey, 1, resetTime, quota)
		So(err, ShouldBeNil)
		So(pass, ShouldBeTrue)
		So(before, ShouldEqual, 0)
		So(after, ShouldEqual, 1)
	})
}

func TestRedisLimitKey(t *testing.T) {
	Convey("redisLimitKey", t, func() {
		appID := config.AppID("test-app")
		limiter := &Limiter{AppID: appID}

		So(limiter.redisLimitKey(model.UsageNameSMS, model.UsageLimitPeriodMonth), ShouldEqual, "app:test-app:usage-limit:SMS")
		So(limiter.redisLimitKey(model.UsageNameSMS, model.UsageLimitPeriodDay), ShouldEqual, "app:test-app:usage-limit:SMS:day")
	})
}

func TestLimiterReserveCountsDayAndMonth(t *testing.T) {
	Convey("Limiter reserves both day and month counters", t, func() {
		ctx := context.Background()
		mr := miniredis.RunT(t)
		now := time.Date(2009, 11, 10, 15, 0, 0, 0, time.UTC)
		mr.SetTime(now)

		pool := redis.NewPool()
		rh := redis.NewHandle(pool, redis.ConnectionOptions{
			RedisURL:              "redis://" + mr.Addr(),
			MaxOpenConnection:     func(i int) *int { return &i }(10),
			MaxIdleConnection:     func(i int) *int { return &i }(5),
			IdleConnectionTimeout: func(d config.DurationSeconds) *config.DurationSeconds { return &d }(300),
			MaxConnectionLifetime: func(d config.DurationSeconds) *config.DurationSeconds { return &d }(900),
		})

		enabled := true
		quota := 10
		limiter := &Limiter{
			Clock: clock.NewMockClockAtTime(now),
			AppID: "test-app",
			Redis: &appredis.Handle{Handle: rh},
			EffectiveConfig: &config.Config{
				FeatureConfig: &config.FeatureConfig{
					Messaging: &config.MessagingFeatureConfig{
						SMSUsage: &config.Deprecated_UsageLimitConfig{
							Enabled: &enabled,
							Period:  config.Deprecated_UsageLimitPeriodMonth,
							Quota:   &quota,
						},
					},
				},
			},
		}

		r, err := limiter.Reserve(ctx, model.UsageNameSMS, 1)
		So(err, ShouldBeNil)
		So(r, ShouldNotBeNil)

		monthly, err := mr.Get("app:test-app:usage-limit:SMS")
		So(err, ShouldBeNil)
		So(monthly, ShouldEqual, "1")

		daily, err := mr.Get("app:test-app:usage-limit:SMS:day")
		So(err, ShouldBeNil)
		So(daily, ShouldEqual, "1")
	})
}

func TestLimiterReserveRollsBackEarlierPeriodOnLaterBlock(t *testing.T) {
	Convey("Limiter rolls back earlier period when a later period blocks", t, func() {
		ctx := context.Background()
		mr := miniredis.RunT(t)
		now := time.Date(2009, 11, 10, 15, 0, 0, 0, time.UTC)
		mr.SetTime(now)

		pool := redis.NewPool()
		rh := redis.NewHandle(pool, redis.ConnectionOptions{
			RedisURL:              "redis://" + mr.Addr(),
			MaxOpenConnection:     func(i int) *int { return &i }(10),
			MaxIdleConnection:     func(i int) *int { return &i }(5),
			IdleConnectionTimeout: func(d config.DurationSeconds) *config.DurationSeconds { return &d }(300),
			MaxConnectionLifetime: func(d config.DurationSeconds) *config.DurationSeconds { return &d }(900),
		})

		limiter := &Limiter{
			Clock: clock.NewMockClockAtTime(now),
			AppID: "test-app",
			Redis: &appredis.Handle{Handle: rh},
			EffectiveConfig: &config.Config{
				FeatureConfig: &config.FeatureConfig{
					Usage: &config.FeatureUsageConfig{
						Limits: &config.FeatureUsageLimitsConfig{
							SMS: []config.FeatureUsageLimitConfig{
								{Quota: 10, Period: model.UsageLimitPeriodDay, Action: model.UsageLimitActionBlock},
								{Quota: 0, Period: model.UsageLimitPeriodMonth, Action: model.UsageLimitActionBlock},
							},
						},
					},
				},
			},
		}

		r, err := limiter.Reserve(ctx, model.UsageNameSMS, 1)
		So(err, ShouldNotBeNil)
		So(r, ShouldBeNil)

		daily, err := mr.Get("app:test-app:usage-limit:SMS:day")
		So(err, ShouldBeNil)
		So(daily, ShouldEqual, "0")
		_, err = mr.Get("app:test-app:usage-limit:SMS")
		So(err, ShouldNotBeNil)
	})
}
