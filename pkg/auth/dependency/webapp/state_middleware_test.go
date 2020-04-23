package webapp

import (
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetPathComponents(t *testing.T) {
	Convey("GetPathComponents", t, func() {
		test := func(uStr string, components ...string) {
			u, err := url.Parse(uStr)
			So(err, ShouldBeNil)
			So(GetPathComponents(u), ShouldResemble, components)
		}
		test("http://example.com")
		test("http://example.com/")
		test("http://example.com/a", "/a")
		test("http://example.com/a/", "/a")
		test("http://example.com/a/b", "/a", "/b")
	})
}
