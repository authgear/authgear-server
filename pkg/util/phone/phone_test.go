package phone

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPhone(t *testing.T) {
	Convey("Phone", t, func() {
		Convey("Mask", func() {
			phone := "+85223456789"
			So(Mask(phone), ShouldEqual, "+8522345****")
		})

		Convey("MaskWithCustomRune", func() {
			phone := "+85223456789"
			So(MaskWithCustomRune(phone, 'x'), ShouldEqual, "+8522345xxxx")
		})
	})
}
