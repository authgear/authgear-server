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

	Convey("ApplyOriginToURL", t, func() {
		test := func(origin string, to string, expected string) {
			originURL, err := url.Parse(origin)
			So(err, ShouldBeNil)

			toURL, err := url.Parse(to)
			So(err, ShouldBeNil)

			actual := ApplyOriginToURL(originURL, toURL)
			So(actual.String(), ShouldEqual, expected)
		}

		test("https://example1.com", "/path", "https://example1.com/path")
		test("https://example1.com", "/path?q=1#frag", "https://example1.com/path?q=1#frag")
		test("https://example1.com:3000", "https://example2.com:3001/path", "https://example1.com:3000/path")
	})
}
