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

	Convey("BestMatch", t, func() {
		test := func(preferred []string, supported []string, expected int) {
			actual, _ := BestMatch(preferred, supported)
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

		// Should select supported tag with higher confidence
		test(
			[]string{"zh-Hant"},
			[]string{"zh-CN", "zh-HK", "en-US"},
			1,
		)
		test(
			[]string{"en-UK"},
			[]string{"zh-CN", "zh-HK", "zh-TW", "en-US"},
			3,
		)
		test(
			[]string{"zh-SG"},
			[]string{"en-US", "zh-CN", "zh-HK", "zh-TW"},
			1,
		)

		// Should select supported tag with lower index if confidence are same
		test(
			[]string{"en"},
			[]string{"en-HK", "en-GB"},
			0,
		)

		// Should select zh-TW with exact confidence
		test(
			[]string{"zh-Hant"},
			[]string{"zh-CN", "zh-HK", "zh-TW", "en-US"},
			2,
		)
		test(
			[]string{"zh-Hant-HK"},
			[]string{"zh-CN", "zh-HK", "zh-TW", "en-US"},
			1,
		)
	})
}
