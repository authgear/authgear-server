package template

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRFC3339(t *testing.T) {
	date := time.Date(2006, 1, 2, 3, 4, 5, 0, time.UTC)

	Convey("RFC3339", t, func() {
		Convey("it supports time.Time", func() {
			So(RFC3339(date), ShouldEqual, "2006-01-02T03:04:05Z")
		})
		Convey("it supports *time.Time", func() {
			So(RFC3339(&date), ShouldEqual, "2006-01-02T03:04:05Z")
		})
		Convey("it does not fail for other data type", func() {
			So(RFC3339(nil), ShouldEqual, "INVALID_DATE")
			So(RFC3339(false), ShouldEqual, "INVALID_DATE")
			So(RFC3339(0), ShouldEqual, "INVALID_DATE")
			So(RFC3339(0.0), ShouldEqual, "INVALID_DATE")
			So(RFC3339(""), ShouldEqual, "INVALID_DATE")
			So(RFC3339(struct{}{}), ShouldEqual, "INVALID_DATE")
			So(RFC3339([]struct{}{}), ShouldEqual, "INVALID_DATE")
		})
	})
}
