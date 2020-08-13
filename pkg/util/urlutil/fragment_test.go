package urlutil

import (
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestWithQueryParamsSetToFragment(t *testing.T) {
	Convey("WithQueryParamsSetToFragment", t, func() {
		test := func(u string, params map[string]string, expected string) {
			uu, err := url.Parse(u)
			So(err, ShouldBeNil)

			actual := WithQueryParamsSetToFragment(uu, params)
			So(actual.String(), ShouldEqual, expected)
		}

		test("http://example.com?a=b", map[string]string{
			"c": "d",
		}, "http://example.com?a=b#c=d")

		test("http://example.com?a=b", map[string]string{
			"c": "#",
		}, "http://example.com?a=b#c=%2523")
	})

}
