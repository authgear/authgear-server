package urlutil

import (
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestExtractOrigin(t *testing.T) {
	Convey("ExtractOrigin", t, func() {
		test := func(u string, expected string) {
			uu, err := url.Parse(u)
			So(err, ShouldBeNil)

			actual := ExtractOrigin(uu)
			So(actual.String(), ShouldEqual, expected)
		}

		test("https://me:pass@example.com:3000/foo/bar?x=1&y=2#anchor", "https://example.com:3000")
		test("http:opaque?x=1&y=2#anchor", "http:opaque")
	})
}
