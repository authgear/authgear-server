package server

import (
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParseListenAddress(t *testing.T) {
	Convey("ParseListenAddress", t, func() {
		test := func(addr string, expected *url.URL) {
			actual, err := ParseListenAddress(addr)
			So(err, ShouldBeNil)
			So(actual, ShouldResemble, expected)
		}

		test("0.0.0.0:3000", &url.URL{
			Scheme: "http",
			Host:   "0.0.0.0:3000",
		})

		test("http://0.0.0.0:3000", &url.URL{
			Scheme: "http",
			Host:   "0.0.0.0:3000",
		})

		test("https://0.0.0.0:3000", &url.URL{
			Scheme: "https",
			Host:   "0.0.0.0:3000",
		})
	})
}
