package errors

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
		So(callerA(0), ShouldStartWith, "errors.callerA")
		So(callerA(1), ShouldStartWith, "errors.TestCaller.")
		So(callerB(0), ShouldStartWith, "errors.callerA")
		So(callerB(1), ShouldStartWith, "errors.callerB")
		So(callerB(2), ShouldStartWith, "errors.TestCaller.")
	})
}
