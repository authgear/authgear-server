package httputil

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestIsJSONContentType(t *testing.T) {
	Convey("IsJSONContentType", t, func() {
		test := func(contentType string, expected bool) {
			So(IsJSONContentType(contentType), ShouldEqual, expected)
		}
		test("text/html", false)
		test("application/json; foo=bar", false)
		test("application/json; charset=utf-8; foobar=bar", false)

		test("application/json", true)
		test("application/json; charset=utf-8", true)
		test("application/json; charset=UTF-8", true)
		test("application/json; charset=uTf-8", true)
	})
}
