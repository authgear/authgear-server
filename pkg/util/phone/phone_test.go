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

func TestIsNorthAmericaNumber(t *testing.T) {
	Convey("IsNorthAmericaNumber", t, func() {
		check := func(e164 string, expected bool, errStr string) {
			actual, err := IsNorthAmericaNumber(e164)
			if errStr == "" {
				So(expected, ShouldEqual, actual)
				So(err, ShouldBeNil)
			} else {
				So(err, ShouldBeError, errStr)
			}
		}

		check("+12015550123", true, "")
		check("+18195555555", true, "")
		check("+61401123456", false, "")
		check("+85298887766", false, "")
		// Possible but invalid number is still a +1 number.
		check("+85212345678", false, "")
		check("+85223456789 ", false, "not in E.164 format")
		check("", false, "not in E.164 format")
	})
}
