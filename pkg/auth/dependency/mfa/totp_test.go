package mfa

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTOTP(t *testing.T) {
	Convey("TOTP", t, func() {
		// nolint: gosec
		fixtureSecret := "GJQFQHET4FX7U5EWSXU36MM36X46TJ7E"
		fixtureTime := time.Date(2019, 6, 1, 0, 0, 0, 0, time.UTC)

		Convey("GenerateTOTPSecret", func() {
			secret, err := GenerateTOTPSecret()
			So(err, ShouldBeNil)
			So(secret, ShouldNotBeEmpty)
			// The secret is of 160 bits
			// Base32 groups 5 bits into 1 character.
			// So the length should be 160/5 = 32.
			So(len(secret), ShouldEqual, 32)
		})

		Convey("GenerateTOTPCode", func() {
			code, err := GenerateTOTPCode(fixtureSecret, fixtureTime)
			So(err, ShouldBeNil)
			// Should be 6 digits
			So(len(code), ShouldEqual, 6)
			So(code, ShouldEqual, "833848")
		})

		Convey("ValidateTOTP", func() {
			Convey("Within the same period", func() {
				code, err := GenerateTOTPCode(fixtureSecret, fixtureTime)
				So(err, ShouldBeNil)

				valid, err := ValidateTOTP(fixtureSecret, code, fixtureTime)
				So(err, ShouldBeNil)
				So(valid, ShouldBeTrue)
			})

			Convey("-1 period", func() {
				code, err := GenerateTOTPCode(fixtureSecret, fixtureTime)
				So(err, ShouldBeNil)

				t1 := fixtureTime.Add(-30 * time.Second)
				t1Code, err := GenerateTOTPCode(fixtureSecret, t1)
				So(err, ShouldBeNil)
				So(t1Code, ShouldNotEqual, code)
				So(t1Code, ShouldEqual, "817861")
				valid, err := ValidateTOTP(fixtureSecret, t1Code, fixtureTime)
				So(err, ShouldBeNil)
				So(valid, ShouldBeTrue)
			})

			Convey("+1 period", func() {
				code, err := GenerateTOTPCode(fixtureSecret, fixtureTime)
				So(err, ShouldBeNil)

				t2 := fixtureTime.Add(30 * time.Second)
				t2Code, err := GenerateTOTPCode(fixtureSecret, t2)
				So(err, ShouldBeNil)
				So(t2Code, ShouldNotEqual, code)
				So(t2Code, ShouldEqual, "503766")
				valid, err := ValidateTOTP(fixtureSecret, t2Code, fixtureTime)
				So(err, ShouldBeNil)
				So(valid, ShouldBeTrue)
			})

			Convey("Invalid code", func() {
				valid, err := ValidateTOTP(fixtureSecret, "123456", fixtureTime)
				So(err, ShouldBeNil)
				So(valid, ShouldBeFalse)
			})

			Convey("Expired code", func() {
				code, err := GenerateTOTPCode(fixtureSecret, fixtureTime)
				So(err, ShouldBeNil)

				t1 := fixtureTime.Add(-60 * time.Second)
				t1Code, err := GenerateTOTPCode(fixtureSecret, t1)
				So(err, ShouldBeNil)
				So(t1Code, ShouldNotEqual, code)
				So(t1Code, ShouldEqual, "369494")
				valid, err := ValidateTOTP(fixtureSecret, t1Code, fixtureTime)
				So(err, ShouldBeNil)
				So(valid, ShouldBeFalse)
			})
		})
	})
}
