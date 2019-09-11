package mail

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParse(t *testing.T) {
	Convey("Parse", t, func() {
		So(Parse("user@example.com"), ShouldBeNil)
		So(Parse("User <user@example.com>"), ShouldBeError, "address must not have name")
		So(Parse(" user@example.com "), ShouldBeError, "formatted address is not the same as input")
		So(Parse("nonsense"), ShouldNotBeNil)
	})
}
