package usage

import (
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	_ "github.com/mattn/go-sqlite3"
	goredis "github.com/redis/go-redis/v9"
	. "github.com/smartystreets/goconvey/convey"

	apievent "github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	dbinfra "github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/otelutil/oteldatabasesql"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

type testEventService struct {
	payloads []apievent.NonBlockingPayload
}

func (s *testEventService) DispatchEventImmediately(ctx context.Context, payload apievent.NonBlockingPayload) error {
	s.payloads = append(s.payloads, payload)
	return nil
}

type testUsageAlertEmailService struct {
	recipients [][]string
	payloads   []*nonblocking.UsageAlertTriggeredEventPayload
	err        error
}

func (s *testUsageAlertEmailService) Send(ctx context.Context, recipients []string, payload *nonblocking.UsageAlertTriggeredEventPayload) error {
	copiedRecipients := append([]string(nil), recipients...)
	s.recipients = append(s.recipients, copiedRecipients)
	s.payloads = append(s.payloads, payload)
	return s.err
}

type testPool struct {
	connPool oteldatabasesql.ConnPool_
}

func (p *testPool) Open(info dbinfra.ConnectionInfo, opts dbinfra.ConnectionOptions) (oteldatabasesql.ConnPool_, error) {
	return p.connPool, nil
}

func (p *testPool) Close() error {
	return nil
}

func newTestAppDBHandle(t *testing.T) *appdb.Handle {
	t.Helper()

	connPool, err := oteldatabasesql.Open(oteldatabasesql.OpenOptions{
		DriverName: "sqlite3",
		DSN:        ":memory:",
	})
	So(err, ShouldBeNil)
	connPool.SetMaxOpenConns(1)
	connPool.SetMaxIdleConns(0)
	connPool.SetConnMaxLifetime(0)
	connPool.SetConnMaxIdleTime(0)

	return &appdb.Handle{
		HookHandle: &dbinfra.HookHandle{
			Pool: &testPool{connPool: connPool},
		},
	}
}

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

func TestLimitRollback(t *testing.T) {
	s := miniredis.RunT(t)

	Convey("rollback tokens", t, func() {
		ctx := context.Background()
		s.FlushAll()

		cli := goredis.NewClient(&goredis.Options{Addr: s.Addr()})
		conn := cli.Conn()

		now := time.UnixMilli(epoch).UTC()
		s.SetTime(now)
		resetTime := time.UnixMilli(epoch).UTC().Add(12 * time.Hour)

		before, after, err := runRollbackScript(ctx, conn, testKey, 1, resetTime)
		So(err, ShouldBeNil)
		So(before, ShouldEqual, 0)
		So(after, ShouldEqual, 0)
		So(s.Exists(testKey), ShouldBeFalse)

		_, _, _, err = runReserveScript(ctx, conn, testKey, 3, resetTime, -1)
		So(err, ShouldBeNil)

		before, after, err = runRollbackScript(ctx, conn, testKey, 1, resetTime)
		So(err, ShouldBeNil)
		So(before, ShouldEqual, 3)
		So(after, ShouldEqual, 2)
		current, err := s.Get(testKey)
		So(err, ShouldBeNil)
		So(current, ShouldEqual, "2")

		before, after, err = runRollbackScript(ctx, conn, testKey, 5, resetTime)
		So(err, ShouldBeNil)
		So(before, ShouldEqual, 2)
		So(after, ShouldEqual, 0)
		So(s.Exists(testKey), ShouldBeFalse)
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
			Clock:    clock.NewMockClockAtTime(now),
			AppID:    "test-app",
			Redis:    &appredis.Handle{Handle: rh},
			Database: newTestAppDBHandle(t),
			EffectiveConfig: &config.Config{
				FeatureConfig: (&config.FeatureConfig{
					Messaging: &config.MessagingFeatureConfig{
						SMSUsage: &config.Deprecated_UsageLimitConfig{
							Enabled: &enabled,
							Period:  config.Deprecated_UsageLimitPeriodMonth,
							Quota:   &quota,
						},
					},
				}).Migrate(),
				AppConfig: &config.AppConfig{},
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

func TestLimiterEffectiveUsageLimitsPrefersUnifiedConfigOverDeprecated(t *testing.T) {
	Convey("Limiter uses unified usage config instead of deprecated feature usage config", t, func() {
		enabled := true
		quota := 10
		limiter := &Limiter{
			EffectiveConfig: &config.Config{
				FeatureConfig: (&config.FeatureConfig{
					Messaging: &config.MessagingFeatureConfig{
						SMSUsage: &config.Deprecated_UsageLimitConfig{
							Enabled: &enabled,
							Period:  config.Deprecated_UsageLimitPeriodMonth,
							Quota:   &quota,
						},
					},
					Usage: &config.FeatureUsageConfig{
						Limits: &config.FeatureUsageLimitsConfig{
							SMS: []config.FeatureUsageLimitConfig{
								{Quota: 3, Period: model.UsageLimitPeriodDay, Action: model.UsageLimitActionAlert},
							},
						},
					},
				}).Migrate(),
				AppConfig: &config.AppConfig{},
			},
		}

		So(limiter.effectiveUsageLimits(model.UsageNameSMS), ShouldResemble, []EffectiveUsageLimit{{
			Name:   model.UsageNameSMS,
			Quota:  3,
			Period: model.UsageLimitPeriodDay,
			Action: model.UsageLimitActionAlert,
		}})
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
			Clock:    clock.NewMockClockAtTime(now),
			AppID:    "test-app",
			Redis:    &appredis.Handle{Handle: rh},
			Database: newTestAppDBHandle(t),
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
				AppConfig: &config.AppConfig{},
			},
		}

		r, err := limiter.Reserve(ctx, model.UsageNameSMS, 1)
		So(err, ShouldNotBeNil)
		So(r, ShouldBeNil)

		_, err = mr.Get("app:test-app:usage-limit:SMS:day")
		So(err, ShouldNotBeNil)
		_, err = mr.Get("app:test-app:usage-limit:SMS")
		So(err, ShouldNotBeNil)
	})
}

func TestLimiterCancelRemovesCounterWhenRollbackReachesZero(t *testing.T) {
	Convey("Limiter cancel removes counter when rollback reaches zero", t, func() {
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
			Clock:    clock.NewMockClockAtTime(now),
			AppID:    "test-app",
			Redis:    &appredis.Handle{Handle: rh},
			Database: newTestAppDBHandle(t),
			EffectiveConfig: &config.Config{
				FeatureConfig: &config.FeatureConfig{
					Usage: &config.FeatureUsageConfig{
						Limits: &config.FeatureUsageLimitsConfig{
							SMS: []config.FeatureUsageLimitConfig{
								{Quota: 10, Period: model.UsageLimitPeriodMonth, Action: model.UsageLimitActionAlert},
							},
						},
					},
				},
				AppConfig: &config.AppConfig{},
			},
		}

		r, err := limiter.Reserve(ctx, model.UsageNameSMS, 1)
		So(err, ShouldBeNil)
		So(r, ShouldNotBeNil)
		current, err := mr.Get("app:test-app:usage-limit:SMS")
		So(err, ShouldBeNil)
		So(current, ShouldEqual, "1")

		limiter.Cancel(ctx, r)

		So(mr.Exists("app:test-app:usage-limit:SMS"), ShouldBeFalse)
	})
}

func TestLimiterDispatchesUsageAlertTriggeredEvent(t *testing.T) {
	Convey("Limiter dispatches usage alert triggered with matched hook URLs", t, func() {
		ctx := context.Background()
		ctx = config.WithAppContext(ctx, &config.AppContext{PlanName: "limited"})
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

		eventService := &testEventService{}
		emailService := &testUsageAlertEmailService{}
		limiter := &Limiter{
			Clock:                  clock.NewMockClockAtTime(now),
			AppID:                  "test-app",
			Redis:                  &appredis.Handle{Handle: rh},
			Database:               newTestAppDBHandle(t),
			EventService:           eventService,
			UsageAlertEmailService: emailService,
			EffectiveConfig: &config.Config{
				FeatureConfig: &config.FeatureConfig{
					Usage: &config.FeatureUsageConfig{
						Hooks: []config.FeatureUsageHookConfig{
							{URL: "https://example.com/sms", Match: "sms"},
							{URL: "https://example.com/all", Match: "*"},
							{URL: "https://example.com/email", Match: "email"},
						},
						Limits: &config.FeatureUsageLimitsConfig{
							SMS: []config.FeatureUsageLimitConfig{
								{Quota: 1, Period: model.UsageLimitPeriodMonth, Action: model.UsageLimitActionAlert},
							},
						},
					},
				},
				AppConfig: &config.AppConfig{
					Usage: &config.UsageConfig{
						Alerts: []config.UsageAlertConfig{
							{Type: "email", Email: "sms@example.com", Match: "sms"},
							{Type: "email", Email: "all@example.com", Match: "*"},
							{Type: "email", Email: "sms@example.com", Match: "*"},
							{Type: "webhook", Email: "ignored@example.com", Match: "sms"},
						},
					},
				},
			},
		}

		r, err := limiter.Reserve(ctx, model.UsageNameSMS, 1)
		So(err, ShouldBeNil)
		So(r, ShouldNotBeNil)
		So(eventService.payloads, ShouldHaveLength, 1)

		payload := eventService.payloads[0].(*nonblocking.UsageAlertTriggeredEventPayload)
		So(payload.Usage.Name, ShouldEqual, model.UsageNameSMS)
		So(payload.Usage.Action, ShouldEqual, model.UsageLimitActionAlert)
		So(payload.Usage.Period, ShouldEqual, model.UsageLimitPeriodMonth)
		So(payload.Usage.CurrentValue, ShouldEqual, 1)
		So(payload.Usage.PlanName, ShouldEqual, "limited")
		So(payload.HookURLs, ShouldResemble, []string{
			"https://example.com/sms",
			"https://example.com/all",
		})
		So(emailService.recipients, ShouldResemble, [][]string{{
			"sms@example.com",
			"all@example.com",
		}})
	})
}

func TestLimiterDoesNotDispatchUsageAlertOnRejectedBlock(t *testing.T) {
	Convey("Limiter does not dispatch duplicate alerts when a block limit rejects without crossing", t, func() {
		var w strings.Builder
		ctx := context.Background()
		ctx = config.WithAppContext(ctx, &config.AppContext{PlanName: "limited"})
		ctx = slogutil.SetContextLogger(ctx, slog.New(slogutil.NewHandlerForTesting(slog.LevelDebug, &w)))
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

		eventService := &testEventService{}
		emailService := &testUsageAlertEmailService{}
		limiter := &Limiter{
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
							SMS: []config.FeatureUsageLimitConfig{
								{Quota: 1, Period: model.UsageLimitPeriodMonth, Action: model.UsageLimitActionBlock},
							},
						},
					},
				},
				AppConfig: &config.AppConfig{
					Usage: &config.UsageConfig{
						Alerts: []config.UsageAlertConfig{
							{Type: "email", Email: "ops@example.com", Match: "*"},
						},
					},
				},
			},
		}

		_, err := limiter.Reserve(ctx, model.UsageNameSMS, 1)
		So(err, ShouldBeNil)
		So(eventService.payloads, ShouldHaveLength, 1)
		So(emailService.recipients, ShouldHaveLength, 1)

		_, err = limiter.Reserve(ctx, model.UsageNameSMS, 1)
		So(err, ShouldNotBeNil)
		So(eventService.payloads, ShouldHaveLength, 1)
		So(emailService.recipients, ShouldHaveLength, 1)
		So(w.String(), ShouldContainSubstring, "usage reservation blocked")
	})
}
