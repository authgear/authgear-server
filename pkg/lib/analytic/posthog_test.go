package analytic

import (
	"encoding/json"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMakeFirstAuthEvents(t *testing.T) {
	Convey("makeFirstAuthEvents", t, func() {
		p := &PosthogIntegration{}
		at := time.Date(2026, 7, 9, 10, 30, 0, 0, time.UTC)

		events, err := p.makeFirstAuthEvents("app-1", map[string]time.Time{
			"client-abc": at,
		})
		So(err, ShouldBeNil)
		So(events, ShouldHaveLength, 1)

		var e map[string]any
		So(json.Unmarshal(events[0], &e), ShouldBeNil)
		So(e["event"], ShouldEqual, "application.first_auth")
		So(e["distinct_id"], ShouldEqual, "client-abc")
		So(e["timestamp"], ShouldEqual, "2026-07-09T10:30:00Z")

		props := e["properties"].(map[string]any)
		So(props["client_id"], ShouldEqual, "client-abc")
		So(props["app_id"], ShouldEqual, "app-1")
		So(props["$process_person_profile"], ShouldEqual, false)

		uuid1 := e["uuid"]
		So(uuid1, ShouldNotBeNil)

		// Determinism: same input -> same uuid (idempotent re-runs).
		events2, err := p.makeFirstAuthEvents("app-1", map[string]time.Time{
			"client-abc": at,
		})
		So(err, ShouldBeNil)
		var e2 map[string]any
		So(json.Unmarshal(events2[0], &e2), ShouldBeNil)
		So(e2["uuid"], ShouldEqual, uuid1)
	})
}
