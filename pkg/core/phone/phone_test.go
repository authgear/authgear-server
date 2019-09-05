package phone

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPhone(t *testing.T) {
	Convey("Phone", t, func() {
		Convey("Parse", func() {
			good := "+85223456789"
			So(Parse(good), ShouldBeNil)

			bad := " +85223456789 "
			So(Parse(bad), ShouldBeError, "not in E.164 format")

			nonsense := "a"
			So(Parse(nonsense), ShouldNotBeNil)
		})

		Convey("Mask", func() {
			phone := "+85223456789"
			So(Mask(phone), ShouldEqual, "+8522345****")
		})
	})
}
