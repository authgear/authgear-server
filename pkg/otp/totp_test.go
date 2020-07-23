package otp_test

import (
	"testing"
	"time"

	"github.com/authgear/authgear-server/pkg/otp"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTOTP(t *testing.T) {
	Convey("TOTP", t, func() {
		// nolint: gosec
		fixtureSecret := "GJQFQHET4FX7U5EWSXU36MM36X46TJ7E"
		fixtureTime := time.Date(2019, 6, 1, 0, 0, 0, 0, time.UTC)

		Convey("GenerateTOTPSecret", func() {
			secret, err := otp.GenerateTOTPSecret()
			So(err, ShouldBeNil)
			So(secret, ShouldNotBeEmpty)
			// The secret is of 160 bits
			// Base32 groups 5 bits into 1 character.
			// So the length should be 160/5 = 32.
			So(len(secret), ShouldEqual, 32)
		})

		Convey("GenerateTOTP", func() {
			code, err := otp.GenerateTOTP(fixtureSecret, fixtureTime, otp.ValidateOptsTOTP)
			So(err, ShouldBeNil)
			// Should be 6 digits
			So(len(code), ShouldEqual, 6)
			So(code, ShouldEqual, "833848")
		})

		Convey("ValidateCode", func() {
			Convey("Within the same period", func() {
				code, err := otp.GenerateTOTP(fixtureSecret, fixtureTime, otp.ValidateOptsTOTP)
				So(err, ShouldBeNil)

				valid := otp.ValidateTOTP(fixtureSecret, code, fixtureTime, otp.ValidateOptsTOTP)
				So(valid, ShouldBeTrue)
			})

			Convey("-1 period", func() {
				code, err := otp.GenerateTOTP(fixtureSecret, fixtureTime, otp.ValidateOptsTOTP)
				So(err, ShouldBeNil)

				t1 := fixtureTime.Add(-30 * time.Second)
				t1Code, err := otp.GenerateTOTP(fixtureSecret, t1, otp.ValidateOptsTOTP)
				So(err, ShouldBeNil)
				So(t1Code, ShouldNotEqual, code)
				So(t1Code, ShouldEqual, "817861")
				valid := otp.ValidateTOTP(fixtureSecret, t1Code, fixtureTime, otp.ValidateOptsTOTP)
				So(valid, ShouldBeTrue)
			})

			Convey("+1 period", func() {
				code, err := otp.GenerateTOTP(fixtureSecret, fixtureTime, otp.ValidateOptsTOTP)
				So(err, ShouldBeNil)

				t2 := fixtureTime.Add(30 * time.Second)
				t2Code, err := otp.GenerateTOTP(fixtureSecret, t2, otp.ValidateOptsTOTP)
				So(err, ShouldBeNil)
				So(t2Code, ShouldNotEqual, code)
				So(t2Code, ShouldEqual, "503766")
				valid := otp.ValidateTOTP(fixtureSecret, t2Code, fixtureTime, otp.ValidateOptsTOTP)
				So(valid, ShouldBeTrue)
			})

			Convey("Invalid code", func() {
				valid := otp.ValidateTOTP(fixtureSecret, "123456", fixtureTime, otp.ValidateOptsTOTP)
				So(valid, ShouldBeFalse)
			})

			Convey("Expired code", func() {
				code, err := otp.GenerateTOTP(fixtureSecret, fixtureTime, otp.ValidateOptsTOTP)
				So(err, ShouldBeNil)

				t1 := fixtureTime.Add(-60 * time.Second)
				t1Code, err := otp.GenerateTOTP(fixtureSecret, t1, otp.ValidateOptsTOTP)
				So(err, ShouldBeNil)
				So(t1Code, ShouldNotEqual, code)
				So(t1Code, ShouldEqual, "369494")
				valid := otp.ValidateTOTP(fixtureSecret, t1Code, fixtureTime, otp.ValidateOptsTOTP)
				So(valid, ShouldBeFalse)
			})
		})
	})
}
