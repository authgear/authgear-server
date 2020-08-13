package intl

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMatch(t *testing.T) {
	Convey("Match", t, func() {
		test := func(preferred []string, supported []string, expected int) {
			actual, _ := Match(preferred, supported)
			So(actual, ShouldEqual, expected)
		}

		// Select default if there is no preferred languages
		test(nil, []string{"en", "ja", "zh"}, 0)
		test([]string{}, []string{"en", "ja", "zh"}, 0)

		// Simply select japanese
		test(
			[]string{"ja-JP", "en-US", "zh-Hant-HK"},
			[]string{"zh", "en", "ja"},
			2,
		)
	})
}
