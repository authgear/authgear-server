package rolesgroupsutil

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFormatKey(t *testing.T) {
	Convey("FormatKey", t, func() {
		f := FormatKey{}.CheckFormat
		So(f(nil), ShouldBeNil)
		So(f(1), ShouldBeNil)
		So(f(""), ShouldBeNil)
		So(f("authgear:"), ShouldBeError, "key cannot start with the preserved prefix: `authgear:`")
	})
}
