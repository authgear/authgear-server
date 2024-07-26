package phone

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLegalAndValidParserCheckE164(t *testing.T) {
	parser := &legalAndValidParser{}
	Convey("LegalAndValidParser.CheckE164", t, func() {
		Convey("good", func() {
			good := "+85223456789"
			So(parser.CheckE164(good), ShouldBeNil)
		})

		Convey("bad", func() {
			bad := " +85223456789 "
			So(parser.CheckE164(bad), ShouldBeError, "not in E.164 format")
		})

		Convey("with letter", func() {
			withLetter := "+85222a"
			So(parser.CheckE164(withLetter), ShouldBeError, "not in E.164 format")
		})

		Convey("Hong Kong phone number does not start with 1", func() {
			invalid := "+85212345678"
			So(parser.CheckE164(invalid), ShouldBeError, "invalid phone number")
		})

		Convey("Emergency phone number", func() {
			emergency := "+852999"
			So(parser.CheckE164(emergency), ShouldBeError, "invalid phone number")
		})

		Convey("1823", func() {
			one_eight_two_three := "+8521823"
			So(parser.CheckE164(one_eight_two_three), ShouldBeError, "invalid phone number")
		})

		Convey("phone number that are relatively new", func() {
			relativelyNew := "+85253580001"
			So(parser.CheckE164(relativelyNew), ShouldBeError, "invalid phone number")
		})

		Convey("too short", func() {
			tooShort := "+85222"
			So(parser.CheckE164(tooShort), ShouldBeError, "invalid phone number")
		})

		Convey("+", func() {
			plus := "+"
			So(parser.CheckE164(plus), ShouldBeError, "not in E.164 format")
		})

		Convey("+country calling code", func() {
			plusCountryCode := "+852"
			So(parser.CheckE164(plusCountryCode), ShouldBeError, "not in E.164 format")
		})

		Convey("letters only", func() {
			nonsense := "a"
			So(parser.CheckE164(nonsense), ShouldBeError, "not in E.164 format")
		})

		Convey("empty", func() {
			empty := ""
			So(parser.CheckE164(empty), ShouldBeError, "not in E.164 format")
		})
	})
}

func TestLegalAndValidParser(t *testing.T) {
	parser := &legalAndValidParser{}
	Convey("Phone", t, func() {
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
	})
}
