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

func TestSupported(t *testing.T) {
	Convey("Supported", t, func() {
		test := func(supported []string, expected []string) {
			actual := Supported(supported, "en")
			So([]string(actual), ShouldResemble, expected)
		}

		test(nil, []string{"en"})
		test([]string{}, []string{"en"})
		test([]string{"ja", "en"}, []string{"en", "ja"})
		test([]string{"ja", "en", "zh"}, []string{"en", "ja", "zh"})
		test([]string{"ja", "zh", "en"}, []string{"en", "ja", "zh"})
		test([]string{"en", "ja", "zh"}, []string{"en", "ja", "zh"})
	})
}
