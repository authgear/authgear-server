package phone

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestIsPhoneNumberCountryAllowed(t *testing.T) {
	Convey("IsPhoneNumberCountryAllowed", t, func() {
		Convey("should return true when allowlist is empty", func() {
			parsed := &ParsedPhoneNumber{
				Alpha2: []string{"HK"},
			}
			result := IsPhoneNumberCountryAllowed(parsed, []string{})
			So(result, ShouldBeTrue)
		})

		Convey("should return true when phone country is in allowlist", func() {
			parsed := &ParsedPhoneNumber{
				Alpha2: []string{"HK"},
			}
			result := IsPhoneNumberCountryAllowed(parsed, []string{"HK", "US", "TW"})
			So(result, ShouldBeTrue)
		})

		Convey("should return false when phone country is not in allowlist", func() {
			parsed := &ParsedPhoneNumber{
				Alpha2: []string{"HK"},
			}
			result := IsPhoneNumberCountryAllowed(parsed, []string{"US", "TW", "GB"})
			So(result, ShouldBeFalse)
		})

		Convey("should return true when any of multiple possible countries is in allowlist", func() {
			parsed := &ParsedPhoneNumber{
				Alpha2: []string{"US", "VI"},
			}
			result := IsPhoneNumberCountryAllowed(parsed, []string{"HK", "VI", "TW"})
			So(result, ShouldBeTrue)
		})

		Convey("should return false when none of multiple possible countries are in allowlist", func() {
			parsed := &ParsedPhoneNumber{
				Alpha2: []string{"US", "VI"},
			}
			result := IsPhoneNumberCountryAllowed(parsed, []string{"HK", "TW", "GB"})
			So(result, ShouldBeFalse)
		})

		Convey("should return true when first possible country is in allowlist", func() {
			parsed := &ParsedPhoneNumber{
				Alpha2: []string{"HK", "MO"},
			}
			result := IsPhoneNumberCountryAllowed(parsed, []string{"HK", "AU"})
			So(result, ShouldBeTrue)
		})

		Convey("should return true when last possible country is in allowlist", func() {
			parsed := &ParsedPhoneNumber{
				Alpha2: []string{"US", "CA", "MO"},
			}
			result := IsPhoneNumberCountryAllowed(parsed, []string{"HK", "MO"})
			So(result, ShouldBeTrue)
		})

		Convey("should return true when middle possible country is in allowlist", func() {
			parsed := &ParsedPhoneNumber{
				Alpha2: []string{"US", "CA", "MO"},
			}
			result := IsPhoneNumberCountryAllowed(parsed, []string{"HK", "CA", "GB"})
			So(result, ShouldBeTrue)
		})

		Convey("should handle case sensitivity correctly", func() {
			parsed := &ParsedPhoneNumber{
				Alpha2: []string{"hk"},
			}
			result := IsPhoneNumberCountryAllowed(parsed, []string{"HK", "US"})
			So(result, ShouldBeFalse)
		})
	})
}
