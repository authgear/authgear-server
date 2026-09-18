package auditlogstreaming

import (
	"context"
	"fmt"
	"testing"

	"github.com/alicebob/miniredis/v2"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/globalredis"
)

// newTestRedisHandle builds a real *globalredis.Handle backed by an
// in-process miniredis server, following the pattern already used for
// Redis-backed code in this repo (pkg/lib/ratelimit, pkg/lib/lockout):
// exercise the Lua against a live instance, because the behaviour under
// test is the script.
func newTestRedisHandle(t *testing.T) *globalredis.Handle {
	t.Helper()
	s := miniredis.RunT(t)

	maxOpen := 10
	maxIdle := 10
	idleTimeout := config.DurationSeconds(60)
	maxLifetime := config.DurationSeconds(60)

	pool := redis.NewPool()
	t.Cleanup(func() { _ = pool.Close() })

	h := redis.NewHandle(pool, redis.ConnectionOptions{
		RedisURL:              "redis://" + s.Addr(),
		MaxOpenConnection:     &maxOpen,
		MaxIdleConnection:     &maxIdle,
		IdleConnectionTimeout: &idleTimeout,
		MaxConnectionLifetime: &maxLifetime,
	})
	return &globalredis.Handle{Handle: h}
}

func TestProducerConsumer(t *testing.T) {
	ctx := context.Background()

	newEvent := func(id string) *event.Event {
		return &event.Event{
			ID:   id,
			Type: "user.authenticated",
			Context: event.Context{
				Timestamp: 1700000000,
				AppID:     "app",
			},
		}
	}

	Convey("Producer and Consumer", t, func() {
		Convey("enqueue then drain returns the entries in the order they were enqueued", func() {
			h := newTestRedisHandle(t)
			p := &Producer{AppID: "app", Streams: []*config.TelemetryAuditLogStreamConfig{{}}, Redis: h}
			c := &Consumer{Redis: h}

			p.Enqueue(ctx, newEvent("one"))
			p.Enqueue(ctx, newEvent("two"))
			p.Enqueue(ctx, newEvent("three"))

			entries, err := c.Drain(ctx, "app")
			So(err, ShouldBeNil)
			So(entries, ShouldHaveLength, 3)
			So(entries[0].Event.ID, ShouldEqual, "one")
			So(entries[1].Event.ID, ShouldEqual, "two")
			So(entries[2].Event.ID, ShouldEqual, "three")
		})

		Convey("drain clears the queue: a second drain returns empty", func() {
			h := newTestRedisHandle(t)
			p := &Producer{AppID: "app", Streams: []*config.TelemetryAuditLogStreamConfig{{}}, Redis: h}
			c := &Consumer{Redis: h}

			p.Enqueue(ctx, newEvent("one"))
			_, err := c.Drain(ctx, "app")
			So(err, ShouldBeNil)

			entries, err := c.Drain(ctx, "app")
			So(err, ShouldBeNil)
			So(entries, ShouldBeEmpty)
		})

		Convey("enqueueing past maxQueueLength drops the oldest, keeping the newest", func() {
			h := newTestRedisHandle(t)
			p := &Producer{AppID: "app", Streams: []*config.TelemetryAuditLogStreamConfig{{}}, Redis: h}
			c := &Consumer{Redis: h}

			for i := range maxQueueLength + 10 {
				p.Enqueue(ctx, newEvent(fmt.Sprintf("entry-%d", i)))
			}

			entries, err := c.Drain(ctx, "app")
			So(err, ShouldBeNil)
			So(entries, ShouldHaveLength, maxQueueLength)
			// The surviving entries are the newest: entries 10..maxQueueLength+9.
			So(entries[0].Event.ID, ShouldEqual, "entry-10")
			So(entries[len(entries)-1].Event.ID, ShouldEqual, fmt.Sprintf("entry-%d", maxQueueLength+9))
		})

		Convey("the pending set contains the app after one enqueue and is empty after ClaimPending", func() {
			h := newTestRedisHandle(t)
			p := &Producer{AppID: "app", Streams: []*config.TelemetryAuditLogStreamConfig{{}}, Redis: h}
			c := &Consumer{Redis: h}

			p.Enqueue(ctx, newEvent("one"))

			appIDs, err := c.ClaimPending(ctx)
			So(err, ShouldBeNil)
			So(appIDs, ShouldContain, "app")

			appIDs, err = c.ClaimPending(ctx)
			So(err, ShouldBeNil)
			So(appIDs, ShouldBeEmpty)
		})

		Convey("ClaimPending returns every app that enqueued, and a re-enqueued app is returned by the next claim", func() {
			h := newTestRedisHandle(t)
			pApp1 := &Producer{AppID: "app1", Streams: []*config.TelemetryAuditLogStreamConfig{{}}, Redis: h}
			pApp2 := &Producer{AppID: "app2", Streams: []*config.TelemetryAuditLogStreamConfig{{}}, Redis: h}
			c := &Consumer{Redis: h}

			pApp1.Enqueue(ctx, newEvent("one"))
			pApp2.Enqueue(ctx, newEvent("one"))

			appIDs, err := c.ClaimPending(ctx)
			So(err, ShouldBeNil)
			So(appIDs, ShouldContain, "app1")
			So(appIDs, ShouldContain, "app2")

			// app1 enqueues again after being claimed.
			pApp1.Enqueue(ctx, newEvent("two"))

			appIDs, err = c.ClaimPending(ctx)
			So(err, ShouldBeNil)
			So(appIDs, ShouldResemble, []string{"app1"})
		})

		Convey("draining an app with no queue returns empty without error", func() {
			h := newTestRedisHandle(t)
			c := &Consumer{Redis: h}

			entries, err := c.Drain(ctx, "no-such-app")
			So(err, ShouldBeNil)
			So(entries, ShouldBeEmpty)
		})

		Convey("both keys carry a TTL after enqueue", func() {
			h := newTestRedisHandle(t)
			p := &Producer{AppID: "app", Streams: []*config.TelemetryAuditLogStreamConfig{{}}, Redis: h}

			p.Enqueue(ctx, newEvent("one"))

			client := h.Client()
			queueTTLResult, ttlErr := client.TTL(ctx, redisKeyQueue("app")).Result()
			So(ttlErr, ShouldBeNil)
			So(queueTTLResult.Seconds(), ShouldBeGreaterThan, 0)

			pendingTTLResult, ttlErr := client.TTL(ctx, redisKeyPending()).Result()
			So(ttlErr, ShouldBeNil)
			So(pendingTTLResult.Seconds(), ShouldBeGreaterThan, 0)
		})

		Convey("a malformed entry is skipped, not the whole batch", func() {
			h := newTestRedisHandle(t)
			p := &Producer{AppID: "app", Streams: []*config.TelemetryAuditLogStreamConfig{{}}, Redis: h}
			c := &Consumer{Redis: h}

			p.Enqueue(ctx, newEvent("one"))
			// drainScript already popped and cleared this by the time
			// decodeQueuedEntry sees it, so there is nothing to retry it
			// from -- Drain must still return the entries that did decode.
			_, err := h.Client().RPush(ctx, redisKeyQueue("app"), "not valid json").Result()
			So(err, ShouldBeNil)
			p.Enqueue(ctx, newEvent("two"))

			entries, err := c.Drain(ctx, "app")
			So(err, ShouldBeNil)
			So(entries, ShouldHaveLength, 2)
			So(entries[0].Event.ID, ShouldEqual, "one")
			So(entries[1].Event.ID, ShouldEqual, "two")
		})

		Convey("disabled or no streams: no Redis call is made", func() {
			h := newTestRedisHandle(t)
			c := &Consumer{Redis: h}

			pDisabled := &Producer{AppID: "app", Streams: []*config.TelemetryAuditLogStreamConfig{{}}, Disabled: true, Redis: h}
			pDisabled.Enqueue(ctx, newEvent("one"))

			pNoStreams := &Producer{AppID: "app", Redis: h}
			pNoStreams.Enqueue(ctx, newEvent("two"))

			entries, err := c.Drain(ctx, "app")
			So(err, ShouldBeNil)
			So(entries, ShouldBeEmpty)
		})
	})
}
