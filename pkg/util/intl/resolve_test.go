package intl

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestResolve(t *testing.T) {
	Convey("Resolve", t, func() {
		test := func(preferred []string, fallback string, supported []string, expected int) {
			actual, _ := Resolve(preferred, fallback, supported)
			So(actual, ShouldEqual, expected)
		}

		// Resolve to default if there no perferred languages
		test(nil, "ja", []string{"en", "ja", "zh"}, 1)
		test([]string{}, "en", []string{"en", "ja", "zh"}, 0)

		// Resolve to default
		test(
			[]string{"ja-JP", "zh-Hant-HK"},
			"en",
			[]string{"af", "en", "hr", "ar"},
			1,
		)
		test(
			[]string{"en-US", "zh-Hant-HK"},
			"en",
			[]string{"zh", "ja"},
			-1,
		)

		// Resolve to japanese
		test(
			[]string{"ja-JP", "en-US", "zh-Hant-HK"},
			"en",
			[]string{"zh", "en", "ja"},
			2,
		)
	})
}
