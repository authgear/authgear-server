package audit_test

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/lib/audit"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/globalredis"
	"github.com/authgear/authgear-server/pkg/lib/telemetry/auditlogstreaming"
)

// fakeNonBlockingPayload is a minimal event.NonBlockingPayload for testing
// Sink.ReceiveNonBlockingEvent's routing, independent of any real payload
// type's other behaviour.
type fakeNonBlockingPayload struct {
	forAudit bool
}

func (p *fakeNonBlockingPayload) UserID() string { return "" }
func (p *fakeNonBlockingPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}
func (p *fakeNonBlockingPayload) FillContext(ctx *event.Context)   {}
func (p *fakeNonBlockingPayload) NonBlockingEventType() event.Type { return "test.fake" }
func (p *fakeNonBlockingPayload) ForHook() bool                    { return true }
func (p *fakeNonBlockingPayload) ForAudit() bool                   { return p.forAudit }
func (p *fakeNonBlockingPayload) RequireReindexUserIDs() []string  { return nil }
func (p *fakeNonBlockingPayload) DeletedUserIDs() []string         { return nil }

func newTestProducer(t *testing.T, appID string) *auditlogstreaming.Producer {
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

	return &auditlogstreaming.Producer{
		AppID:   config.AppID(appID),
		Streams: []*config.TelemetryAuditLogStreamConfig{{}},
		Redis:   &globalredis.Handle{Handle: h},
	}
}

func queueLength(t *testing.T, p *auditlogstreaming.Producer, appID string) int {
	t.Helper()
	c := &auditlogstreaming.Consumer{Redis: p.Redis}
	entries, err := c.Drain(context.Background(), appID)
	if err != nil {
		t.Fatal(err)
	}
	return len(entries)
}

func newTestEvent(payload event.NonBlockingPayload) *event.Event {
	return &event.Event{
		ID:            "test-id",
		Type:          payload.NonBlockingEventType(),
		Payload:       payload,
		Context:       event.Context{Timestamp: 1700000000, AppID: "app"},
		IsNonBlocking: true,
	}
}

func TestSinkReceiveNonBlockingEvent(t *testing.T) {
	ctx := context.Background()

	Convey("Sink.ReceiveNonBlockingEvent", t, func() {
		Convey("ForAudit() == false: neither persisted nor enqueued", func() {
			producer := newTestProducer(t, "app")
			// Database is deliberately nil: if the sink proceeded past the
			// ForAudit() check instead of returning immediately, the next
			// thing it does is call s.Database.WithTx, which would panic
			// on a nil *auditdb.WriteHandle -- a loud, unambiguous signal
			// that the routing is wrong, rather than a silent one.
			sink := &audit.Sink{Database: nil, Producer: producer}

			e := newTestEvent(&fakeNonBlockingPayload{forAudit: false})
			err := sink.ReceiveNonBlockingEvent(ctx, e)
			So(err, ShouldBeNil)

			So(queueLength(t, producer, "app"), ShouldEqual, 0)
		})

		Convey("Database == nil: not enqueued", func() {
			producer := newTestProducer(t, "app")
			sink := &audit.Sink{Database: nil, Producer: producer}

			e := newTestEvent(&fakeNonBlockingPayload{forAudit: true})
			err := sink.ReceiveNonBlockingEvent(ctx, e)
			So(err, ShouldBeNil)

			So(queueLength(t, producer, "app"), ShouldEqual, 0)
		})
	})
}
