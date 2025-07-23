package phone

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/nyaruka/phonenumbers"
	. "github.com/smartystreets/goconvey/convey"
)

// As of 2025-07-23, this is a reserved number.
// See https://www.ofca.gov.hk/filemanager/ofca/en/content_311/no_plan.pdf
//
// In case you update "github.com/nyaruka/phonenumbers", and found that
// RESERVED_NUMBER became assigned (not reserved),
// you need to find another number that is marked as "** Reserved 預留" in the PDF.
const RESERVED_NUMBER = "+85253530000"

func TestParsePhoneNumberWithUserInput(t *testing.T) {

	Convey("ParsePhoneNumberWithUserInput", t, func() {
		Convey("Good Hong Kong number", func() {
			good := "+85223456789"

			parsed, err := ParsePhoneNumberWithUserInput(good)
			So(err, ShouldBeNil)
			So(parsed.E164, ShouldEqual, "+85223456789")
			So(parsed.Alpha2, ShouldEqual, []string{"HK"})
			So(parsed.IsPossibleNumber, ShouldBeTrue)
			So(parsed.IsValidNumber, ShouldBeTrue)
			So(parsed.CountryCallingCodeWithoutPlusSign, ShouldEqual, "852")
			So(parsed.NationalNumberWithoutFormatting, ShouldEqual, "23456789")
		})

		Convey("Good Australia number", func() {
			good := "+61401123456"

			parsed, err := ParsePhoneNumberWithUserInput(good)
			So(err, ShouldBeNil)
			So(parsed.E164, ShouldEqual, "+61401123456")
			So(parsed.Alpha2, ShouldEqual, []string{"AU"})
			So(parsed.IsPossibleNumber, ShouldBeTrue)
			So(parsed.IsValidNumber, ShouldBeTrue)
			So(parsed.CountryCallingCodeWithoutPlusSign, ShouldEqual, "61")
			So(parsed.NationalNumberWithoutFormatting, ShouldEqual, "401123456")
		})

		Convey("Good United States Virgin Islands phone number", func() {
			// http://www.wtng.info/wtng-1340-vi.html
			// This website says 712 xxxx is valid.
			good := "+13407121234"

			parsed, err := ParsePhoneNumberWithUserInput(good)
			So(err, ShouldBeNil)
			So(parsed.E164, ShouldEqual, "+13407121234")
			So(parsed.Alpha2, ShouldEqual, []string{"VI"})
			So(parsed.IsPossibleNumber, ShouldBeTrue)
			So(parsed.IsValidNumber, ShouldBeTrue)
			So(parsed.CountryCallingCodeWithoutPlusSign, ShouldEqual, "1")
			So(parsed.NationalNumberWithoutFormatting, ShouldEqual, "3407121234")
		})

		Convey("Good British Virgin Islands phone number", func() {
			good := "+12841234567"

			parsed, err := ParsePhoneNumberWithUserInput(good)
			So(err, ShouldBeNil)
			So(parsed.E164, ShouldEqual, "+12841234567")
			So(parsed.Alpha2, ShouldEqual, []string{"VG"})
			So(parsed.IsPossibleNumber, ShouldBeTrue)
			// I cannot find a valid pattern on the Internet.
			So(parsed.IsValidNumber, ShouldBeFalse)
			So(parsed.CountryCallingCodeWithoutPlusSign, ShouldEqual, "1")
			So(parsed.NationalNumberWithoutFormatting, ShouldEqual, "2841234567")
		})

		Convey("Good Isle of Man phone number", func() {
			good := "+447624123456"

			parsed, err := ParsePhoneNumberWithUserInput(good)
			So(err, ShouldBeNil)
			So(parsed.E164, ShouldEqual, "+447624123456")
			So(parsed.Alpha2, ShouldEqual, []string{"IM"})
			So(parsed.IsPossibleNumber, ShouldBeTrue)
			So(parsed.IsValidNumber, ShouldBeTrue)
			So(parsed.CountryCallingCodeWithoutPlusSign, ShouldEqual, "44")
			So(parsed.NationalNumberWithoutFormatting, ShouldEqual, "7624123456")
		})

		Convey("Not in E164", func() {
			bad := " +85223456789 "

			parsed, err := ParsePhoneNumberWithUserInput(bad)
			So(err, ShouldBeNil)
			So(parsed.IsPossibleNumber, ShouldBeTrue)
			So(parsed.IsValidNumber, ShouldBeTrue)
			So(parsed.RequireUserInputInE164(), ShouldBeError, "not in E.164 format")
		})

		Convey("with letter", func() {
			withLetter := "+85222a"

			_, err := ParsePhoneNumberWithUserInput(withLetter)
			So(err, ShouldBeError, "not in E.164 format")
		})

		Convey("Hong Kong phone number does not start with 1", func() {
			invalid := "+85212345678"

			parsed, err := ParsePhoneNumberWithUserInput(invalid)
			So(err, ShouldBeNil)
			So(parsed.IsPossibleNumber, ShouldBeTrue)
			So(parsed.IsValidNumber, ShouldBeFalse)
			So(parsed.Alpha2, ShouldEqual, []string{"HK"})
		})

		Convey("Emergency phone number", func() {
			emergency := "+852999"

			parsed, err := ParsePhoneNumberWithUserInput(emergency)
			So(err, ShouldBeNil)
			So(parsed.IsPossibleNumber, ShouldBeFalse)
			So(parsed.IsValidNumber, ShouldBeFalse)
			So(parsed.Alpha2, ShouldEqual, []string{"HK"})
		})

		Convey("1823", func() {
			one_eight_two_three := "+8521823"

			parsed, err := ParsePhoneNumberWithUserInput(one_eight_two_three)
			So(err, ShouldBeNil)
			So(parsed.IsPossibleNumber, ShouldBeFalse)
			So(parsed.IsValidNumber, ShouldBeFalse)
			So(parsed.Alpha2, ShouldEqual, []string{"HK"})
		})

		Convey("reserved number is not valid", func() {
			parsed, err := ParsePhoneNumberWithUserInput(RESERVED_NUMBER)
			So(err, ShouldBeNil)
			So(parsed.IsPossibleNumber, ShouldBeTrue)
			So(parsed.IsValidNumber, ShouldBeFalse)
			So(parsed.Alpha2, ShouldEqual, []string{"HK"})
		})

		Convey("too short", func() {
			tooShort := "+85222"

			parsed, err := ParsePhoneNumberWithUserInput(tooShort)
			So(err, ShouldBeNil)
			So(parsed.IsPossibleNumber, ShouldBeFalse)
			So(parsed.IsValidNumber, ShouldBeFalse)
			So(parsed.Alpha2, ShouldEqual, []string{"HK"})
		})

		Convey("+", func() {
			plus := "+"

			_, err := ParsePhoneNumberWithUserInput(plus)
			So(err, ShouldBeError, "not in E.164 format")
		})

		Convey("+country calling code", func() {
			plusCountryCode := "+852"

			_, err := ParsePhoneNumberWithUserInput(plusCountryCode)
			So(err, ShouldBeError, "not in E.164 format")
		})

		Convey("letters only", func() {
			nonsense := "a"

			_, err := ParsePhoneNumberWithUserInput(nonsense)
			So(err, ShouldBeError, "not in E.164 format")
		})

		Convey("empty", func() {
			empty := ""

			_, err := ParsePhoneNumberWithUserInput(empty)
			So(err, ShouldBeError, "not in E.164 format")
		})
	})
}

func TestIsNorthAmericaNumber(t *testing.T) {
	Convey("IsNorthAmericaNumber", t, func() {
		check := func(e164 string, expected bool, errStr string) {
			parsed, err := ParsePhoneNumberWithUserInput(e164)
			if errStr == "" {
				So(err, ShouldBeNil)
				So(parsed.IsNorthAmericaNumber(), ShouldEqual, expected)
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
		check("+85223456789 ", false, "")
		check("", false, "not in E.164 format")
	})
}

func TestRequire_IsPossibleNumber_IsValidNumber_UserInputInE164(t *testing.T) {
	Convey("Require_IsPossibleNumber_IsValidNumber_UserInputInE164", t, func() {
		test := func(input string, expectedErrorStr string) {
			err := Require_IsPossibleNumber_IsValidNumber_UserInputInE164(input)
			if expectedErrorStr == "" {
				So(err, ShouldBeNil)
			} else {
				So(err, ShouldBeError, expectedErrorStr)
			}
		}

		// good
		test("+85298765432", "")
		// Reserved number is not IsValidNumber.
		test(RESERVED_NUMBER, "invalid phone number")
		// Not In E164
		test(" +85298765432", "not in E.164 format")
	})
}

func TestParse_IsPossibleNumber_ReturnE164(t *testing.T) {
	Convey("Parse_IsPossibleNumber_ReturnE164", t, func() {
		test := func(input string, expectedErrorStr string) {
			_, err := Parse_IsPossibleNumber_ReturnE164(input)
			if expectedErrorStr == "" {
				So(err, ShouldBeNil)
			} else {
				So(err, ShouldBeError, expectedErrorStr)
			}
		}

		// good
		test("+85298765432", "")
		// relatively new number is IsPossibleNumber
		test("+85253580001", "")
		// Return E164 even if the input is not originally in E164.
		test(" +852 9876 5432 ", "")
	})
}

func TestCountryCallingCodeToRegions(t *testing.T) {
	Convey("Country calling codes to regions", t, func() {
		c := phonenumbers.GetSupportedCallingCodes()
		m := map[int][]string{}
		for countryCallingCode := range c {
			codes := phonenumbers.GetRegionCodesForCountryCode(countryCallingCode)
			m[countryCallingCode] = codes
		}
		data, err := os.ReadFile("calling_codes_regions.json")
		if err != nil {
			panic(err)
		}

		var expectedMap map[int][]string

		err = json.Unmarshal(data, &expectedMap)
		if err != nil {
			panic(err)
		}

		So(m, ShouldEqual, expectedMap)
	})
}
