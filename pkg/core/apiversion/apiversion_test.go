package apiversion

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFormat(t *testing.T) {
	Convey("Format", t, func() {
		So(Format(1, 2), ShouldEqual, "v1.2")
	})
}

func TestParse(t *testing.T) {
	Convey("Parse", t, func() {
		var major, minor int
		var ok bool

		major, minor, ok = Parse("v1.2")
		So(ok, ShouldBeTrue)
		So(major, ShouldEqual, 1)
		So(minor, ShouldEqual, 2)

		_, _, ok = Parse("v1.2 ")
		So(ok, ShouldBeFalse)
	})
}
