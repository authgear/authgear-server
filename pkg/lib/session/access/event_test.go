package access

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewAccessEvent(t *testing.T) {
	Convey("NewAccessEvent", t, func() {
		now := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		Convey("should record current timestamp", func() {
			event := NewEvent(now, "", "")
			So(event.Timestamp, ShouldResemble, now)
		})
		Convey("should populate connection info", func() {
			event := NewEvent(now, "216.58.197.110", "")
			So(event.RemoteIP, ShouldResemble, "216.58.197.110")
		})
		Convey("should populate user agent", func() {
			event := NewEvent(now, "", "SDK")
			So(event.UserAgent, ShouldEqual, "SDK")
		})
	})
}
