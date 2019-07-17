package url

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestQuery(t *testing.T) {
	Convey("Test Query", t, func() {
		Convey("should encode empty query to empty string", func() {
			q := Query{}
			So(q.Encode(), ShouldEqual, "")
		})
		Convey("should encode single param", func() {
			q := Query{}
			q.Add("a", "b")
			So(q.Encode(), ShouldEqual, "a=b")
		})
		Convey("should encode empty value", func() {
			q := Query{}
			q.Add("a", "")
			So(q.Encode(), ShouldEqual, "a=")
		})
		Convey("should encode many params", func() {
			q := Query{}
			q.Add("a", "b")
			q.Add("c", "")
			q.Add("a", "d")
			So(q.Encode(), ShouldEqual, "a=b&c=&a=d")
		})
		Convey("should escape both key and value", func() {
			q := Query{}
			q.Add("我", "你")
			So(q.Encode(), ShouldEqual, "%E6%88%91=%E4%BD%A0")
		})
		Convey("should escape space into +", func() {
			q := Query{}
			q.Add("a", "hello world")
			So(q.Encode(), ShouldEqual, "a=hello+world")
		})
	})
}
