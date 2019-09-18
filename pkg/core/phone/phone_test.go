package phone

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPhone(t *testing.T) {
	Convey("Phone", t, func() {
		Convey("EnsureE164", func() {
			good := "+85223456789"
			So(EnsureE164(good), ShouldBeNil)

			bad := " +85223456789 "
			So(EnsureE164(bad), ShouldBeError, "not in E.164 format")

			nonsense := "a"
			So(EnsureE164(nonsense), ShouldNotBeNil)
		})

		Convey("Mask", func() {
			phone := "+85223456789"
			So(Mask(phone), ShouldEqual, "+8522345****")
		})
	})
}
