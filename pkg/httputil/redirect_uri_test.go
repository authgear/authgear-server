package httputil

import (
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestHostRelative(t *testing.T) {
	Convey("HostRelative", t, func() {
		test := func(input string, expected string) {
			u, err := url.Parse(input)
			So(err, ShouldBeNil)
			actual := HostRelative(u)
			So(actual.String(), ShouldEqual, expected)
		}

		test("http://example.com", "/")
		test("http://example.com/", "/")
		test("http://example.com/a", "/a")
		test("http://example.com/a?a", "/a?a")
		test("http://example.com/a?a#a", "/a?a#a")
	})
}
