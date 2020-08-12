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
			check := func(nationalNumber, callingCode, e164 string) {
				actual, err := Parse(nationalNumber, callingCode)
				if e164 == "" {
					So(err, ShouldNotBeNil)
				} else {
					So(actual, ShouldEqual, e164)
				}
			}
			// calling code can have optional + sign
			check("99887766", "+852", "+85299887766")
			check("99887766", "852", "+85299887766")

			// national number can have spaces in between
			check("9988 7766", "852", "+85299887766")
			check(" 9988 7766 ", "852", "+85299887766")

			// national number can have hyphens in between
			check("9988-7766", "852", "+85299887766")
			check(" 9988-7766 ", "852", "+85299887766")
			check("99-88-77-66", "852", "+85299887766")
			check("9-9-8-8-7-7-6-6", "852", "+85299887766")
			check("9 - 9 - 8 - 8 - 7 - 7 - 6 - 6 ", "852", "+85299887766")

			// calling code can have leading or trailing spaces
			check("99887766", " +852 ", "+85299887766")
			check("99887766", " 852 ", "+85299887766")

			// calling code can have spaces or hyphens in between
			check("99887766", "8 52", "+85299887766")
			check("99887766", "8-52", "+85299887766")
			check("99887766", " 8-5-2- ", "+85299887766")
		})

		Convey("Mask", func() {
			phone := "+85223456789"
			So(Mask(phone), ShouldEqual, "+8522345****")
		})
	})
}
