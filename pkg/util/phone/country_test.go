package phone

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPackageConstants(t *testing.T) {
	Convey("package constants", t, func() {
		Convey("has at least 10 entries", func() {
			So(len(AllCountries), ShouldBeGreaterThan, 10)
		})

		Convey("contains US", func() {
			var us *Country
			for _, country := range AllCountries {
				if country.Alpha2 == "US" {
					c := country
					us = &c
				}
			}
			So(us, ShouldNotBeNil)
			So(us.CountryCallingCode, ShouldEqual, "1")
		})

		Convey("AllCountries and AllAlpha2 have the same length", func() {
			So(len(AllCountries), ShouldEqual, len(AllAlpha2))
		})
	})
}
