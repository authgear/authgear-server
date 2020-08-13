package log

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParseLevel(t *testing.T) {
	Convey("ParseLevel", t, func() {
		debug, err := ParseLevel("Debug")
		So(err, ShouldBeNil)
		So(debug, ShouldEqual, LevelDebug)

		info, err := ParseLevel("inFO")
		So(err, ShouldBeNil)
		So(info, ShouldEqual, LevelInfo)

		warn, err := ParseLevel("warn")
		So(err, ShouldBeNil)
		So(warn, ShouldEqual, LevelWarn)

		warning, err := ParseLevel("WARNING")
		So(err, ShouldBeNil)
		So(warning, ShouldEqual, LevelWarn)

		errorLevel, err := ParseLevel("error")
		So(err, ShouldBeNil)
		So(errorLevel, ShouldEqual, LevelError)

		_, err = ParseLevel("foobar")
		So(err, ShouldBeError, "log: unknown level: foobar")
	})
}
