package configsource_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
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
			escaped := configsource.EscapePath(testCase.raw)
			So(escaped, ShouldEqual, testCase.escaped)
			raw, err := configsource.UnescapePath(escaped)
			So(err, ShouldBeNil)
			So(raw, ShouldEqual, testCase.raw)
		}
	})
}
