package validation

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFormatBCP47(t *testing.T) {
	f := FormatBCP47{}.CheckFormat

	Convey("FormatBCP47", t, func() {
		So(f(1), ShouldBeNil)

		So(f(""), ShouldBeError, "invalid BCP 47 tag: language: tag is not well-formed")
		So(f("a"), ShouldBeError, "invalid BCP 47 tag: language: tag is not well-formed")
		So(f("foobar"), ShouldBeError, "invalid BCP 47 tag: language: tag is not well-formed")

		So(f("en"), ShouldBeNil)
		So(f("zh-TW"), ShouldBeNil)
		So(f("und"), ShouldBeNil)
	})
}
