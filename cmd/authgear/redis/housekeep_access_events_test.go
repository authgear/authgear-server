package redis

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestHousekeepTaskAccessEvents(t *testing.T) {
	Convey("access-events", t, func() {
		f := newHousekeepFixture(t)

		// Deleted: every stream, whether or not it has an expiry.
		f.xadd("app:accounts:access-events:session-no-expiry")
		f.xadd("app:accounts:access-events:session-with-expiry")
		f.expire("app:accounts:access-events:session-with-expiry", time.Hour)

		// Left alone: not a stream.
		f.set("app:accounts:access-events:not-a-stream", "value", 0)
		// Left alone: not an access events key, though the name contains access-events.
		f.xadd("app:accounts:session:access-events:a")

		stdout, _ := f.run(false, housekeepTaskAccessEvents)
		So(stdout, ShouldEqual, `deleted app:accounts:access-events:session-no-expiry
deleted app:accounts:access-events:session-with-expiry
`)
		So(f.memoryRedis.Keys(), ShouldResemble, []string{
			"app:accounts:access-events:not-a-stream",
			"app:accounts:session:access-events:a",
		})
	})
}
