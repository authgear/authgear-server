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

		Convey("Parse", func() {
			Convey("valid cases", func() {
				check := func(nationalNumber, callingCode, e164 string) {
					actual, err := Parse(nationalNumber, callingCode)
					if e164 == "" {
						So(err, ShouldNotBeNil)
					} else {
						So(actual, ShouldEqual, e164)
					}
				}
				// calling code can have optional + sign
				check("98887766", "+852", "+85298887766")
				check("98887766", "852", "+85298887766")

				// national number can have spaces in between
				check("9888 7766", "852", "+85298887766")
				check(" 9888 7766 ", "852", "+85298887766")

				// national number can have hyphens in between
				check("9888-7766", "852", "+85298887766")
				check(" 9888-7766 ", "852", "+85298887766")
				check("98-88-77-66", "852", "+85298887766")
				check("9-8-8-8-7-7-6-6", "852", "+85298887766")
				check("9 - 8 - 8 - 8 - 7 - 7 - 6 - 6 ", "852", "+85298887766")

				// calling code can have leading or trailing spaces
				check("98887766", " +852 ", "+85298887766")
				check("98887766", " 852 ", "+85298887766")

				// calling code can have spaces or hyphens in between
				check("98887766", "8 52", "+85298887766")
				check("98887766", "8-52", "+85298887766")
				check("98887766", " 8-5-2- ", "+85298887766")
			})

			Convey("should not accept non-numeric character(s)", func() {
				_, err := Parse("asdf", "+852")
				So(err, ShouldBeError, "not in E.164 format")
			})

			Convey("invalid, (+852) phone number does not begin with 1", func() {
				_, err := Parse("12345678", "+852")
				So(err, ShouldBeError, "invalid phone number")
			})

			Convey("invalid, (+852) phone number is 8 digit long", func() {
				_, err := Parse("6234567", "+852")
				So(err, ShouldBeError, "invalid phone number")
			})
		})

		Convey("Mask", func() {
			phone := "+85223456789"
			So(Mask(phone), ShouldEqual, "+8522345****")
		})
	})
}
