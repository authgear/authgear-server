package labelutil

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLabel(t *testing.T) {
	Convey("Label", t, func() {
		So(Label("a"), ShouldEqual, "A")
		So(Label("a_pen"), ShouldEqual, "A Pen")
		So(Label("foobar"), ShouldEqual, "Foobar")
		So(Label("a_to_b"), ShouldEqual, "A to B")
		So(Label("a_b_c_d"), ShouldEqual, "A B C D")
	})
}
