package phone

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLegalAndValidParser(t *testing.T) {
	parser := &legalAndValidParser{}
	Convey("Phone", t, func() {
		Convey("checkE164", func() {
			good := "+85223456789"
			So(parser.CheckE164(good), ShouldBeNil)

			bad := " +85223456789 "
			So(parser.CheckE164(bad), ShouldBeError, "not in E.164 format")

			withLetter := "+85222a"
			So(parser.CheckE164(withLetter), ShouldBeError, "not in E.164 format")

			invalid := "+85212345678"
			So(parser.CheckE164(invalid), ShouldBeError, "invalid phone number")

			tooShort := "+85222"
			So(parser.CheckE164(tooShort), ShouldBeError, "invalid phone number")

			plus := "+"
			So(parser.CheckE164(plus), ShouldBeError, "not in E.164 format")

			plusCountryCode := "+852"
			So(parser.CheckE164(plusCountryCode), ShouldBeError, "not in E.164 format")

			nonsense := "a"
			So(parser.CheckE164(nonsense), ShouldBeError, "not in E.164 format")

			empty := ""
			So(parser.CheckE164(empty), ShouldBeError, "not in E.164 format")

			// Hong Kong number starting with 7
			So(parser.CheckE164("+85270123456"), ShouldBeNil)
		})

		Convey("ParseCountryCallingCodeAndNationalNumber", func() {
			Convey("valid cases", func() {
				check := func(nationalNumber, callingCode, e164 string) {
					actual, err := parser.ParseCountryCallingCodeAndNationalNumber(nationalNumber, callingCode)
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
				_, err := parser.ParseCountryCallingCodeAndNationalNumber("asdf", "+852")
				So(err, ShouldBeError, "not in E.164 format")
			})

			Convey("invalid, (+852) phone number does not begin with 1", func() {
				_, err := parser.ParseCountryCallingCodeAndNationalNumber("12345678", "+852")
				So(err, ShouldBeError, "invalid phone number")
			})

			Convey("invalid, (+852) phone number is 8 digit long", func() {
				_, err := parser.ParseCountryCallingCodeAndNationalNumber("6234567", "+852")
				So(err, ShouldBeError, "invalid phone number")
			})
		})

		Convey("SplitE164", func() {
			Convey("valid cases", func() {
				check := func(e164, nationalNumber, callingCode string) {
					actualNationalNumber, actualCallingCode, err := parser.SplitE164(e164)
					So(actualNationalNumber, ShouldEqual, nationalNumber)
					So(actualCallingCode, ShouldEqual, callingCode)
					So(err, ShouldBeNil)
				}

				check("+61401123456", "401123456", "61")
				check("+85298887766", "98887766", "852")
			})

			Convey("invalid, not in E.164 format", func() {
				check := func(input string) {
					_, _, err := parser.SplitE164(input)
					So(err, ShouldBeError, "not in E.164 format")
				}

				check("unknown")
				check("85298887766")
			})

			Convey("invalid, invalid phone number", func() {
				check := func(input string) {
					_, _, err := parser.SplitE164(input)
					So(err, ShouldBeError, "invalid phone number")
				}

				check("+85212345678")
				check("+852123456")
			})
		})

		Convey("IsNorthAmericaNumber", func() {
			check := func(e164 string, expected bool, errStr string) {
				actual, err := parser.IsNorthAmericaNumber(e164)
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
			check("+85212345678", false, "invalid phone number")
			check("+85223456789 ", false, "not in E.164 format")
			check("", false, "not in E.164 format")
		})
	})
}
