package filepathutil

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPathEscape(t *testing.T) {
	Convey("Path escaping", t, func() {
		cases := []struct {
			raw     string
			escaped string
		}{
			{"", ""},
			{"test.html", "test.html"},
			{"foo/test.html", "foo_2f_test.html"},
			{"../bar/test.html", ".._2f_bar_2f_test.html"},
			{"zh_hk/auth-header.html", "zh_5f_hk_2f_auth-header.html"},
			{"$$config", "_24__24_config"},
		}

		for _, testCase := range cases {
			escaped := EscapePath(testCase.raw)
			So(escaped, ShouldEqual, testCase.escaped)
			raw, err := UnescapePath(escaped)
			So(err, ShouldBeNil)
			So(raw, ShouldEqual, testCase.raw)
		}
	})
}
