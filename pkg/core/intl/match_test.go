package intl

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSortSupported(t *testing.T) {
	Convey("SortSupported", t, func() {
		test := func(supported []string, expected []string) {
			actual := SortSupported(supported, "en")
			So(actual, ShouldResemble, expected)
		}

		test(nil, []string{"en"})
		test([]string{}, []string{"en"})
		test([]string{"ja", "en"}, []string{"en", "ja"})
		test([]string{"ja", "en", "zh"}, []string{"en", "ja", "zh"})
		test([]string{"ja", "zh", "en"}, []string{"en", "ja", "zh"})
		test([]string{"en", "ja", "zh"}, []string{"en", "ja", "zh"})
	})
}

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
