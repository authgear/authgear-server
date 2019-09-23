package mail

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMaskAddress(t *testing.T) {
	Convey("MaskAddress", t, func() {
		So(MaskAddress("user@example.com"), ShouldEqual, "us**@example.com")
		So(MaskAddress("johndoe@example.com"), ShouldEqual, "joh****@example.com")
	})
}
