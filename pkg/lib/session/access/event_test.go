package access

import (
	"net/http"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewAccessEvent(t *testing.T) {
	Convey("NewAccessEvent", t, func() {
		now := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		Convey("should record current timestamp", func() {
			req, _ := http.NewRequest("POST", "", nil)

			event := NewEvent(now, req, true)
			So(event.Timestamp, ShouldResemble, now)
		})
		Convey("should populate connection info", func() {
			req, _ := http.NewRequest("POST", "", nil)
			req.RemoteAddr = "192.168.1.11:31035"
			req.Header.Set("X-Forwarded-For", "13.225.103.28, 216.58.197.110")
			req.Header.Set("X-Real-IP", "216.58.197.110")
			req.Header.Set("Forwarded", "for=216.58.197.110;proto=http;by=192.168.1.11")

			event := NewEvent(now, req, true)
			So(event.RemoteIP, ShouldResemble, "216.58.197.110")
			event = NewEvent(now, req, false)
			So(event.RemoteIP, ShouldResemble, "192.168.1.11")
		})
		Convey("should populate user agent", func() {
			req, _ := http.NewRequest("POST", "", nil)
			req.RemoteAddr = "192.168.1.11:31035"
			req.Header.Set("User-Agent", "SDK")

			event := NewEvent(now, req, true)
			So(event.UserAgent, ShouldEqual, "SDK")
		})
	})
}
