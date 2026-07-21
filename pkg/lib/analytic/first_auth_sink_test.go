package analytic

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

func TestBuildFirstAuthEvent(t *testing.T) {
	Convey("buildFirstAuthEvent", t, func() {
		at := time.Date(2026, 7, 9, 10, 30, 0, 0, time.UTC)

		b, err := buildFirstAuthEvent("app-1", "client-abc", at)
		So(err, ShouldBeNil)

		var e map[string]any
		So(json.Unmarshal(b, &e), ShouldBeNil)
		So(e["event"], ShouldEqual, "application.first_auth")
		So(e["distinct_id"], ShouldEqual, "client-abc")
		So(e["timestamp"], ShouldEqual, "2026-07-09T10:30:00Z")

		props := e["properties"].(map[string]any)
		So(props["client_id"], ShouldEqual, "client-abc")
		So(props["app_id"], ShouldEqual, "app-1")
		So(props["$geoip_disable"], ShouldEqual, true)
		So(props["$process_person_profile"], ShouldEqual, false)

		Convey("uuid is deterministic per (app_id, client_id)", func() {
			b2, err := buildFirstAuthEvent("app-1", "client-abc", at.Add(time.Hour))
			So(err, ShouldBeNil)
			var e2 map[string]any
			So(json.Unmarshal(b2, &e2), ShouldBeNil)
			So(e2["uuid"], ShouldEqual, e["uuid"])
		})

		Convey("uuid differs for a different client", func() {
			b3, err := buildFirstAuthEvent("app-1", "client-xyz", at)
			So(err, ShouldBeNil)
			var e3 map[string]any
			So(json.Unmarshal(b3, &e3), ShouldBeNil)
			So(e3["uuid"], ShouldNotEqual, e["uuid"])
		})
	})
}

func TestFirstAuthDedupKey(t *testing.T) {
	Convey("firstAuthDedupKey", t, func() {
		So(firstAuthDedupKey("app-1", "client-abc"), ShouldEqual, "app:app-1:posthog-first-auth:client-abc")
	})
}

func TestFirstAuthSinkNoop(t *testing.T) {
	Convey("FirstAuthSink.ReceiveNonBlockingEvent no-op paths", t, func() {
		ctx := context.Background()
		// AnalyticRedis is nil and credentials are nil; every path below must
		// short-circuit before touching Redis, so no panic and a nil error.
		sink := &FirstAuthSink{
			Clock:         clock.NewMockClockAt("2026-07-09T10:30:00Z"),
			AnalyticRedis: nil,
			Posthog:       &PosthogService{PosthogCredentials: nil},
		}

		Convey("ignores non-auth events", func() {
			e := &event.Event{
				Type:    event.Type("user.created"),
				Context: event.Context{AppID: "app-1", ClientID: "client-abc"},
			}
			So(sink.ReceiveNonBlockingEvent(ctx, e), ShouldBeNil)
		})

		Convey("ignores auth events with no client_id", func() {
			e := &event.Event{
				Type:    nonblocking.UserAuthenticated,
				Context: event.Context{AppID: "app-1", ClientID: ""},
			}
			So(sink.ReceiveNonBlockingEvent(ctx, e), ShouldBeNil)
		})

		Convey("no-op when PostHog credentials are unset", func() {
			e := &event.Event{
				Type:    nonblocking.UserAuthenticated,
				Context: event.Context{AppID: "app-1", ClientID: "client-abc"},
			}
			So(sink.ReceiveNonBlockingEvent(ctx, e), ShouldBeNil)
		})

		Convey("handles both auth event types in the filter", func() {
			for _, t := range []event.Type{nonblocking.UserAuthenticated, nonblocking.M2MTokenCreated} {
				e := &event.Event{
					Type:    t,
					Context: event.Context{AppID: "app-1", ClientID: "client-abc"},
				}
				So(sink.ReceiveNonBlockingEvent(ctx, e), ShouldBeNil)
			}
		})

		Convey("blocking events are ignored", func() {
			e := &event.Event{
				Type:    nonblocking.UserAuthenticated,
				Context: event.Context{AppID: "app-1", ClientID: "client-abc"},
			}
			So(sink.ReceiveBlockingEvent(ctx, e), ShouldBeNil)
		})
	})
}
