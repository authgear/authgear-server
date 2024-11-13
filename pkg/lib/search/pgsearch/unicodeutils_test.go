package pgsearch

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnicodeSegmentation(t *testing.T) {
	Convey("StringUnicodeSegmentation", t, func() {
		test := func(input string, expectedOutput string) {
			output := StringUnicodeSegmentation(input)
			So(output, ShouldEqual, expectedOutput)
		}

		test("", "")
		test("test string", "test   string")
		test("測試", "測 試")
		test("テスト", "テスト")
		test("試す", "試 す")
	})

	Convey("MapUnicodeSegmentation", t, func() {
		input := map[string]string{
			"":            "",
			"test string": "test string",
			"測試":          "測試",
			"テスト":         "テスト",
			"試す":          "試す",
		}
		output := MapUnicodeSegmentation(input)

		So(output, ShouldEqual, map[string]string{
			"":            "",
			"test string": "test   string",
			"測試":          "測 試",
			"テスト":         "テスト",
			"試す":          "試 す",
		})
	})
}
