package password

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConfirmPassword(t *testing.T) {
	Convey("ConfirmPassword", t, func() {
		So(ConfirmPassword("a", "a"), ShouldBeNil)
		So(ConfirmPassword("", ""), ShouldBeNil)

		So(ConfirmPassword("a", ""), ShouldNotBeNil)
		So(ConfirmPassword("", "a"), ShouldNotBeNil)
		So(ConfirmPassword("a", "b"), ShouldNotBeNil)
	})
}
