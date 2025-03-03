package errorutil

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestForceLogging(t *testing.T) {
	Convey("ForceLogging and IsForceLogging", t, func() {
		Convey("nil error is not force logging", func() {
			So(IsForceLogging(nil), ShouldBeFalse)
		})

		Convey("plain error is not force logging", func() {
			So(IsForceLogging(fmt.Errorf("a")), ShouldBeFalse)
		})

		Convey("cannot force logging nil error", func() {
			So(IsForceLogging(ForceLogging(nil)), ShouldBeFalse)
		})

		Convey("can force logging any error", func() {
			So(IsForceLogging(ForceLogging(fmt.Errorf("a"))), ShouldBeTrue)
		})

		Convey("calling ForceLogging more than once is the same as calling it once", func() {
			So(IsForceLogging(ForceLogging(ForceLogging(fmt.Errorf("a")))), ShouldBeTrue)
		})
	})
}
