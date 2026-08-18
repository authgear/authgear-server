package usage

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

func newTestStandingLimiter(t *testing.T, eventService *testEventService, emailService *testUsageAlertEmailService, limits []config.StandingFeatureUsageLimitConfig, alerts []config.UsageAlertConfig) *Limiter {
	t.Helper()

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

	return &Limiter{
		Clock:                  clock.NewMockClockAtTime(now),
		AppID:                  "test-app",
		Redis:                  &appredis.Handle{Handle: rh},
		Database:               newTestAppDBHandle(t),
		EventService:           eventService,
		UsageAlertEmailService: emailService,
		EffectiveConfig: &config.Config{
			FeatureConfig: &config.FeatureConfig{
				Usage: &config.FeatureUsageConfig{
					Limits: &config.FeatureUsageLimitsConfig{
						OAuthClientDCR: limits,
					},
				},
			},
			AppConfig: &config.AppConfig{
				Usage: &config.UsageConfig{
					Alerts: alerts,
				},
			},
		},
	}
}

func TestLimiterCheckStanding(t *testing.T) {
	Convey("Limiter.CheckStanding", t, func() {
		ctx := context.Background()

		Convey("no configured limit always returns nil", func() {
			limiter := newTestStandingLimiter(t, &testEventService{}, &testUsageAlertEmailService{}, nil, nil)
			err := limiter.CheckStanding(ctx, model.UsageNameOAuthClientDCR, 1000)
			So(err, ShouldBeNil)
		})

		Convey("action: alert only, never blocks regardless of count", func() {
			limiter := newTestStandingLimiter(t, &testEventService{}, &testUsageAlertEmailService{}, []config.StandingFeatureUsageLimitConfig{
				{Quota: 20, Action: model.UsageLimitActionAlert},
			}, nil)
			err := limiter.CheckStanding(ctx, model.UsageNameOAuthClientDCR, 1000)
			So(err, ShouldBeNil)
		})

		Convey("action: block, quota: 20", func() {
			limiter := newTestStandingLimiter(t, &testEventService{}, &testUsageAlertEmailService{}, []config.StandingFeatureUsageLimitConfig{
				{Quota: 20, Action: model.UsageLimitActionBlock},
			}, nil)

			Convey("currentCount well below quota: nil", func() {
				err := limiter.CheckStanding(ctx, model.UsageNameOAuthClientDCR, 5)
				So(err, ShouldBeNil)
			})

			Convey("currentCount one below quota: nil (creating the 20th client reaches exactly the quota, still allowed)", func() {
				err := limiter.CheckStanding(ctx, model.UsageNameOAuthClientDCR, 19)
				So(err, ShouldBeNil)
			})

			Convey("currentCount at or above quota: ErrStandingUsageLimitExceeded", func() {
				err := limiter.CheckStanding(ctx, model.UsageNameOAuthClientDCR, 20)
				So(err, ShouldNotBeNil)
			})
		})

		Convey("the strictest (minimum) block quota applies when multiple block entries exist", func() {
			limiter := newTestStandingLimiter(t, &testEventService{}, &testUsageAlertEmailService{}, []config.StandingFeatureUsageLimitConfig{
				{Quota: 20, Action: model.UsageLimitActionBlock},
				{Quota: 5, Action: model.UsageLimitActionBlock},
			}, nil)
			So(limiter.CheckStanding(ctx, model.UsageNameOAuthClientDCR, 4), ShouldBeNil)
			So(limiter.CheckStanding(ctx, model.UsageNameOAuthClientDCR, 5), ShouldNotBeNil)
		})
	})
}

func TestLimiterReportStandingCreated(t *testing.T) {
	Convey("Limiter.ReportStandingCreated", t, func() {
		ctx := context.Background()
		ctx = config.WithAppContext(ctx, &config.AppContext{PlanName: "limited"})

		Convey("fires once when countBeforeCreate+1 crosses a configured quota", func() {
			eventService := &testEventService{}
			emailService := &testUsageAlertEmailService{}
			limiter := newTestStandingLimiter(t, eventService, emailService, []config.StandingFeatureUsageLimitConfig{
				{Quota: 2, Action: model.UsageLimitActionBlock},
			}, []config.UsageAlertConfig{
				{Type: "email", Email: "admin@example.com", Match: "oauth_client_dcr"},
			})

			limiter.ReportStandingCreated(ctx, model.UsageNameOAuthClientDCR, 1)

			So(eventService.payloads, ShouldHaveLength, 1)
			payload := eventService.payloads[0].(*nonblocking.UsageAlertTriggeredEventPayload)
			So(payload.Usage.Name, ShouldEqual, model.UsageNameOAuthClientDCR)
			So(payload.Usage.Action, ShouldEqual, model.UsageLimitActionBlock)
			So(payload.Usage.Period, ShouldEqual, model.UsageLimitPeriod(""))
			So(payload.Usage.Quota, ShouldEqual, 2)
			So(payload.Usage.CurrentValue, ShouldEqual, 2)
			So(emailService.recipients, ShouldResemble, [][]string{{"admin@example.com"}})
		})

		Convey("does not fire when countBeforeCreate+1 does not cross any quota", func() {
			eventService := &testEventService{}
			emailService := &testUsageAlertEmailService{}
			limiter := newTestStandingLimiter(t, eventService, emailService, []config.StandingFeatureUsageLimitConfig{
				{Quota: 2, Action: model.UsageLimitActionBlock},
			}, nil)

			limiter.ReportStandingCreated(ctx, model.UsageNameOAuthClientDCR, 5)

			So(eventService.payloads, ShouldHaveLength, 0)
		})

		Convey("does not fire again on a repeated call with the same already-crossed count", func() {
			eventService := &testEventService{}
			emailService := &testUsageAlertEmailService{}
			limiter := newTestStandingLimiter(t, eventService, emailService, []config.StandingFeatureUsageLimitConfig{
				{Quota: 2, Action: model.UsageLimitActionBlock},
			}, nil)

			limiter.ReportStandingCreated(ctx, model.UsageNameOAuthClientDCR, 1)
			So(eventService.payloads, ShouldHaveLength, 1)

			// A second registration, now that the quota was already crossed
			// (countBeforeCreate=2 -> after=3), must not fire the same
			// quota's alert again -- it already crossed at count 2.
			limiter.ReportStandingCreated(ctx, model.UsageNameOAuthClientDCR, 2)
			So(eventService.payloads, ShouldHaveLength, 1)
		})
	})
}
