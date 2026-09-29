package auditlogstreaming

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type fakeSender struct {
	appCtx  *config.AppContext
	entries []QueuedEntry
}

func (s *fakeSender) Send(ctx context.Context, entries []QueuedEntry) {
	s.entries = entries
}

type fakeSenderFactory struct {
	mu      sync.Mutex
	senders map[string]*fakeSender
}

func newFakeSenderFactory() *fakeSenderFactory {
	return &fakeSenderFactory{senders: map[string]*fakeSender{}}
}

func (f *fakeSenderFactory) MakeSender(appID string, appCtx *config.AppContext) Sender {
	s := &fakeSender{appCtx: appCtx}
	f.mu.Lock()
	f.senders[appID] = s
	f.mu.Unlock()
	return s
}

// trackingSender calls onSend synchronously from within Send, letting a
// test observe how many are in flight at once (and block them to keep
// them in flight for the observation).
type trackingSender struct {
	onSend func()
}

func (s *trackingSender) Send(ctx context.Context, entries []QueuedEntry) {
	s.onSend()
}

type trackingSenderFactory struct {
	onSend func()
}

func (f *trackingSenderFactory) MakeSender(appID string, appCtx *config.AppContext) Sender {
	return &trackingSender{onSend: f.onSend}
}

type fakeAppContextResolver struct {
	contexts map[string]*config.AppContext
}

func (r *fakeAppContextResolver) ResolveContext(ctx context.Context, appID string, fn func(context.Context, *config.AppContext) error) error {
	appCtx, ok := r.contexts[appID]
	if !ok {
		return fmt.Errorf("no such app: %s", appID)
	}
	return fn(ctx, appCtx)
}

// newTestAppContext builds a minimal *config.AppContext with the fields
// Runnable.drain and NewSenderImpl actually read.
func newTestAppContext(publicOrigin string, streams []*config.TelemetryAuditLogStreamConfig, disabled bool) *config.AppContext {
	return &config.AppContext{
		Config: &config.Config{
			AppConfig: &config.AppConfig{
				HTTP: &config.HTTPConfig{PublicOrigin: publicOrigin},
				Telemetry: &config.TelemetryConfig{
					AuditLogs: &config.TelemetryAuditLogsConfig{Streams: streams},
				},
			},
			FeatureConfig: &config.FeatureConfig{
				Telemetry: &config.TelemetryFeatureConfig{
					AuditLogs: &config.TelemetryAuditLogsFeatureConfig{
						Streaming: &config.TelemetryAuditLogsStreamingFeatureConfig{
							Disabled: &disabled,
						},
					},
				},
			},
			SecretConfig: &config.SecretConfig{},
		},
	}
}

func TestRunnableRun(t *testing.T) {
	ctx := context.Background()

	newEvent := func(id string) *event.Event {
		return &event.Event{ID: id, Type: "user.authenticated", Context: event.Context{Timestamp: 1700000000}}
	}

	Convey("Runnable.Run", t, func() {
		Convey("two apps queued: each app's batch is sent with that app's own appCtx", func() {
			h := newTestRedisHandle(t)

			p1 := &Producer{AppID: "app1", Streams: []*config.TelemetryAuditLogStreamConfig{{}}, Redis: h, Interval: config.DurationString("1m")}
			p1.Enqueue(ctx, newEvent("one"))
			p2 := &Producer{AppID: "app2", Streams: []*config.TelemetryAuditLogStreamConfig{{}}, Redis: h, Interval: config.DurationString("1m")}
			p2.Enqueue(ctx, newEvent("two"))

			streams := []*config.TelemetryAuditLogStreamConfig{{Name: "collector"}}
			resolver := &fakeAppContextResolver{contexts: map[string]*config.AppContext{
				"app1": newTestAppContext("http://app1.example.com", streams, false),
				"app2": newTestAppContext("http://app2.example.com", streams, false),
			}}
			senderFactory := newFakeSenderFactory()

			r := &Runnable{
				Consumer:           &Consumer{Redis: h},
				AppContextResolver: resolver,
				SenderFactory:      senderFactory,
				Redis:              h,
				Interval:           config.DurationString("1m"),
			}

			err := r.Run(ctx)
			So(err, ShouldBeNil)

			So(senderFactory.senders["app1"], ShouldNotBeNil)
			So(senderFactory.senders["app1"].entries, ShouldHaveLength, 1)
			So(senderFactory.senders["app1"].appCtx.Config.AppConfig.HTTP.PublicOrigin, ShouldEqual, "http://app1.example.com")

			So(senderFactory.senders["app2"], ShouldNotBeNil)
			So(senderFactory.senders["app2"].entries, ShouldHaveLength, 1)
			So(senderFactory.senders["app2"].appCtx.Config.AppConfig.HTTP.PublicOrigin, ShouldEqual, "http://app2.example.com")
		})

		Convey("one app's resolve failure does not prevent the other's delivery", func() {
			h := newTestRedisHandle(t)

			p1 := &Producer{AppID: "app1", Streams: []*config.TelemetryAuditLogStreamConfig{{}}, Redis: h, Interval: config.DurationString("1m")}
			p1.Enqueue(ctx, newEvent("one"))
			p2 := &Producer{AppID: "app2", Streams: []*config.TelemetryAuditLogStreamConfig{{}}, Redis: h, Interval: config.DurationString("1m")}
			p2.Enqueue(ctx, newEvent("two"))

			streams := []*config.TelemetryAuditLogStreamConfig{{Name: "collector"}}
			// app1 is deliberately absent from the resolver, simulating a
			// resolve failure.
			resolver := &fakeAppContextResolver{contexts: map[string]*config.AppContext{
				"app2": newTestAppContext("http://app2.example.com", streams, false),
			}}
			senderFactory := newFakeSenderFactory()

			r := &Runnable{
				Consumer:           &Consumer{Redis: h},
				AppContextResolver: resolver,
				SenderFactory:      senderFactory,
				Redis:              h,
				Interval:           config.DurationString("1m"),
			}

			err := r.Run(ctx)
			So(err, ShouldBeNil)

			So(senderFactory.senders["app1"], ShouldBeNil)
			So(senderFactory.senders["app2"], ShouldNotBeNil)
			So(senderFactory.senders["app2"].entries, ShouldHaveLength, 1)
		})

		Convey("an app whose feature config is disabled at drain time delivers nothing", func() {
			h := newTestRedisHandle(t)

			p := &Producer{AppID: "app", Streams: []*config.TelemetryAuditLogStreamConfig{{}}, Redis: h, Interval: config.DurationString("1m")}
			p.Enqueue(ctx, newEvent("one"))

			streams := []*config.TelemetryAuditLogStreamConfig{{Name: "collector"}}
			resolver := &fakeAppContextResolver{contexts: map[string]*config.AppContext{
				// disabled=true, even though the entry was enqueued while
				// the Producer thought streaming was enabled.
				"app": newTestAppContext("http://app.example.com", streams, true),
			}}
			senderFactory := newFakeSenderFactory()

			r := &Runnable{
				Consumer:           &Consumer{Redis: h},
				AppContextResolver: resolver,
				SenderFactory:      senderFactory,
				Redis:              h,
				Interval:           config.DurationString("1m"),
			}

			err := r.Run(ctx)
			So(err, ShouldBeNil)
			So(senderFactory.senders["app"], ShouldBeNil)
		})

		Convey("an app whose streams were removed between enqueue and drain delivers nothing and does not error", func() {
			h := newTestRedisHandle(t)

			p := &Producer{AppID: "app", Streams: []*config.TelemetryAuditLogStreamConfig{{}}, Redis: h, Interval: config.DurationString("1m")}
			p.Enqueue(ctx, newEvent("one"))

			resolver := &fakeAppContextResolver{contexts: map[string]*config.AppContext{
				"app": newTestAppContext("http://app.example.com", nil, false),
			}}
			senderFactory := newFakeSenderFactory()

			r := &Runnable{
				Consumer:           &Consumer{Redis: h},
				AppContextResolver: resolver,
				SenderFactory:      senderFactory,
				Redis:              h,
				Interval:           config.DurationString("1m"),
			}

			err := r.Run(ctx)
			So(err, ShouldBeNil)
			So(senderFactory.senders["app"], ShouldBeNil)
		})

		Convey("the redsync mutex is held for the tick: a concurrent Run returns without draining", func() {
			h := newTestRedisHandle(t)

			p := &Producer{AppID: "app", Streams: []*config.TelemetryAuditLogStreamConfig{{}}, Redis: h, Interval: config.DurationString("1m")}
			p.Enqueue(ctx, newEvent("one"))

			resolver := &fakeAppContextResolver{contexts: map[string]*config.AppContext{}}
			senderFactory := newFakeSenderFactory()

			r := &Runnable{
				Consumer:           &Consumer{Redis: h},
				AppContextResolver: resolver,
				SenderFactory:      senderFactory,
				Redis:              h,
				Interval:           config.DurationString("1m"),
			}

			// Simulate another replica already draining, by holding the
			// same-named lock directly.
			mutex := h.NewMutex(drainMutexName)
			So(mutex.LockContext(ctx), ShouldBeNil)
			defer func() { _, _ = mutex.UnlockContext(context.Background()) }()

			err := r.Run(ctx)
			So(err, ShouldBeNil)
			So(senderFactory.senders, ShouldBeEmpty)

			// The queue was never claimed, let alone drained.
			c := &Consumer{Redis: h}
			entries, err := c.Drain(ctx, "app")
			So(err, ShouldBeNil)
			So(entries, ShouldHaveLength, 1)
		})

		Convey("a failed Drain re-marks the app pending instead of losing it until the queue expires", func() {
			h := newTestRedisHandle(t)

			p := &Producer{AppID: "app", Streams: []*config.TelemetryAuditLogStreamConfig{{}}, Redis: h, Interval: config.DurationString("1m")}
			p.Enqueue(ctx, newEvent("one"))

			// Force drainScript's LRANGE/DEL to fail with a genuine
			// Redis error (WRONGTYPE) without touching the connection,
			// so the rest of the tick runs normally and this exercises
			// the real MarkPending call, not a stub.
			client := h.Client()
			So(client.Del(ctx, redisKeyQueue("app")).Err(), ShouldBeNil)
			So(client.Set(ctx, redisKeyQueue("app"), "not-a-list", 0).Err(), ShouldBeNil)

			resolver := &fakeAppContextResolver{contexts: map[string]*config.AppContext{}}
			senderFactory := newFakeSenderFactory()

			r := &Runnable{
				Consumer:           &Consumer{Redis: h},
				AppContextResolver: resolver,
				SenderFactory:      senderFactory,
				Redis:              h,
				Interval:           config.DurationString("1m"),
			}

			err := r.Run(ctx)
			So(err, ShouldBeNil)
			So(senderFactory.senders, ShouldBeEmpty)

			// ClaimPending removed "app" from the pending set before
			// Drain failed; MarkPending must have put it back, or the
			// entry below becomes unreachable until the TTL expires it.
			members, err := client.SMembers(ctx, redisKeyPending()).Result()
			So(err, ShouldBeNil)
			So(members, ShouldContain, "app")

			// The queue itself was untouched by the failed Drain: fix
			// the WRONGTYPE value and confirm the original entry is
			// still there, reachable on the next tick.
			So(client.Del(ctx, redisKeyQueue("app")).Err(), ShouldBeNil)
			p2 := &Producer{AppID: "app", Streams: []*config.TelemetryAuditLogStreamConfig{{}}, Redis: h, Interval: config.DurationString("1m")}
			p2.Enqueue(ctx, newEvent("two"))

			c := &Consumer{Redis: h}
			entries, err := c.Drain(ctx, "app")
			So(err, ShouldBeNil)
			So(entries, ShouldHaveLength, 1)
			So(entries[0].Event.ID, ShouldEqual, "two")
		})

		Convey("apps are drained concurrently, bounded by maxConcurrentAppDrains", func() {
			h := newTestRedisHandle(t)

			n := maxConcurrentAppDrains + 5
			streams := []*config.TelemetryAuditLogStreamConfig{{Name: "collector"}}
			contexts := map[string]*config.AppContext{}
			for i := range n {
				appID := fmt.Sprintf("app%d", i)
				p := &Producer{AppID: config.AppID(appID), Streams: streams, Redis: h, Interval: config.DurationString("1m")}
				p.Enqueue(ctx, newEvent("one"))
				contexts[appID] = newTestAppContext(fmt.Sprintf("http://%s.example.com", appID), streams, false)
			}

			var mu sync.Mutex
			current, peak := 0, 0
			release := make(chan struct{})
			senderFactory := &trackingSenderFactory{onSend: func() {
				mu.Lock()
				current++
				if current > peak {
					peak = current
				}
				mu.Unlock()

				<-release

				mu.Lock()
				current--
				mu.Unlock()
			}}

			r := &Runnable{
				Consumer:           &Consumer{Redis: h},
				AppContextResolver: &fakeAppContextResolver{contexts: contexts},
				SenderFactory:      senderFactory,
				Redis:              h,
				Interval:           config.DurationString("1m"),
			}

			runDone := make(chan error, 1)
			go func() { runDone <- r.Run(ctx) }()

			// Let every goroutine that is going to start, start and block
			// in onSend, then unblock them all at once.
			time.Sleep(300 * time.Millisecond)
			close(release)
			err := <-runDone
			So(err, ShouldBeNil)

			mu.Lock()
			defer mu.Unlock()
			// Concurrency actually happened (not silently serialized)...
			So(peak, ShouldBeGreaterThan, 1)
			// ...but never more than the bound, even with more apps
			// queued than the bound.
			So(peak, ShouldBeLessThanOrEqualTo, maxConcurrentAppDrains)
		})
	})
}
