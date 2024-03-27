package intl

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMatch(t *testing.T) {
	Convey("Match", t, func() {
		test := func(preferred []string, supported []string, expected int) {
			actual, _ := Match_Deprecated(preferred, supported)
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

		// Simply select zh-HK
		test(
			[]string{"zh-hk"},
			[]string{"zh-TW", "zh-HK", "en-US"},
			1,
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
		// Should select zh-HK with exact confidence
		test(
			[]string{"zh-Hant-HK"},
			[]string{"zh-CN", "zh-HK", "zh-TW", "en-US"},
			1,
		)

		// Should use preceding preferred language if possible
		test(
			[]string{"ja-JP", "en-GB", "zh-HK"},
			[]string{"zh-HK", "zh-TW", "en-US"},
			2,
		)
	})
}

type BenchmarkTestFixture struct {
	Preferred []string
	Supported []string
}

var benchmarkTestFixtures []*BenchmarkTestFixture = []*BenchmarkTestFixture{
	{
		Preferred: []string{},
		Supported: []string{"en", "ja", "zh"},
	},
	{
		Preferred: []string{"ja-JP", "en-US", "zh-Hant-HK"},
		Supported: []string{"zh", "en", "ja"},
	},
	{
		Preferred: []string{"zh-hk"},
		Supported: []string{"zh-TW", "zh-HK", "en-US"},
	},
	{
		Preferred: []string{"zh-Hant"},
		Supported: []string{"zh-CN", "zh-HK", "zh-TW", "en-US"},
	},
	{
		Preferred: []string{"zh-HK", "zh-Hant-HK", "zh-TW", "zh"},
		Supported: []string{"zh-CN", "zh-HK", "zh-TW", "zh-SG", "zh-MC", "en-US", "en-GB", "ja-JP"},
	},
}

func BenchmarkMatch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, f := range benchmarkTestFixtures {
			Match_Deprecated(f.Preferred, f.Supported)
		}
	}
}

func BenchmarkBestMatch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, f := range benchmarkTestFixtures {
			BestMatch(f.Preferred, f.Supported)
		}
	}
}
