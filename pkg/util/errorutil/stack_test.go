package errorutil

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func callerA(skip int) string {
	return stack(skip)
}

func callerB(skip int) string {
	return callerA(skip)
}

func TestCaller(t *testing.T) {
	Convey("caller", t, func() {
		So(callerA(0), ShouldStartWith, "errorutil.callerA")
		So(callerA(1), ShouldStartWith, "errorutil.TestCaller.")
		So(callerB(0), ShouldStartWith, "errorutil.callerA")
		So(callerB(1), ShouldStartWith, "errorutil.callerB")
		So(callerB(2), ShouldStartWith, "errorutil.TestCaller.")
	})
}
