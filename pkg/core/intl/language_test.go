package intl

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFallback(t *testing.T) {
	Convey("Fallback", t, func() {
		So(Fallback("ja"), ShouldEqual, "ja")
		So(Fallback(""), ShouldEqual, DefaultLanguage)
	})
}
