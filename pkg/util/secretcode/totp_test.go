package secretcode

import (
	"encoding/base32"
	"testing"
	"time"

	coreimage "github.com/authgear/authgear-server/pkg/util/image"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTOTP(t *testing.T) {
	Convey("totp", t, func() {
		// nolint: gosec
		fixtureSecret := "GJQFQHET4FX7U5EWSXU36MM36X46TJ7E"
		fixtureTime := time.Date(2019, 6, 1, 0, 0, 0, 0, time.UTC)
		totp := NewTOTPFromSecret(fixtureSecret)

		Convey("NewTOTPSecretFromRNG", func() {
			totp, err := NewTOTPFromRNG()
			So(err, ShouldBeNil)
			So(totp.Secret, ShouldNotBeEmpty)
			// The secret is of 160 bits
			// Base32 groups 5 bits into 1 character.
			// So the length should be 160/5 = 32.
			So(len(totp.Secret), ShouldEqual, 32)
		})

		Convey("GenerateCode", func() {
			code, err := totp.GenerateCode(fixtureTime)
			So(err, ShouldBeNil)
			// Should be 6 digits
			So(len(code), ShouldEqual, 6)
			So(code, ShouldEqual, "833848")
		})

		Convey("ValidateCode", func() {
			Convey("Within the same period", func() {
				code, err := totp.GenerateCode(fixtureTime)
				So(err, ShouldBeNil)

				valid := totp.ValidateCode(fixtureTime, code)
				So(valid, ShouldBeTrue)
			})

			Convey("-1 period", func() {
				code, err := totp.GenerateCode(fixtureTime)
				So(err, ShouldBeNil)

				t1 := fixtureTime.Add(-30 * time.Second)
				t1Code, err := totp.GenerateCode(t1)
				So(err, ShouldBeNil)
				So(t1Code, ShouldNotEqual, code)
				So(t1Code, ShouldEqual, "817861")
				valid := totp.ValidateCode(fixtureTime, t1Code)
				So(valid, ShouldBeTrue)
			})

			Convey("+1 period", func() {
				code, err := totp.GenerateCode(fixtureTime)
				So(err, ShouldBeNil)

				t2 := fixtureTime.Add(30 * time.Second)
				t2Code, err := totp.GenerateCode(t2)
				So(err, ShouldBeNil)
				So(t2Code, ShouldNotEqual, code)
				So(t2Code, ShouldEqual, "503766")
				valid := totp.ValidateCode(fixtureTime, t2Code)
				So(valid, ShouldBeTrue)
			})

			Convey("Invalid code", func() {
				valid := totp.ValidateCode(fixtureTime, "123456")
				So(valid, ShouldBeFalse)
			})

			Convey("Expired code", func() {
				code, err := totp.GenerateCode(fixtureTime)
				So(err, ShouldBeNil)

				t1 := fixtureTime.Add(-60 * time.Second)
				t1Code, err := totp.GenerateCode(t1)
				So(err, ShouldBeNil)
				So(t1Code, ShouldNotEqual, code)
				So(t1Code, ShouldEqual, "369494")
				valid := totp.ValidateCode(fixtureTime, t1Code)
				So(valid, ShouldBeFalse)
			})
		})
	})
}

func TestMakeTOTPKey(t *testing.T) {
	Convey("MakeTOTPKey", t, func() {
		// Use a fixed secret to make the test stable.
		// This must be at least 20 bytes.
		rawSecret := "01234567890123456789"
		enc := base32.StdEncoding.WithPadding(base32.NoPadding)
		secret := enc.EncodeToString([]byte(rawSecret))
		totp := NewTOTPFromSecret(secret)

		opts := QRCodeImageOptions{
			Issuer:      "test",
			AccountName: "john.doe@example.com",
			Width:       100,
			Height:      100,
		}

		img, err := totp.QRCodeImage(opts)
		So(err, ShouldBeNil)

		dataURI, err := coreimage.DataURIFromImage(coreimage.CodecPNG, img)
		So(err, ShouldBeNil)
		// Copy the data URI in your browser and you should see an QR code image.
		// The image has been verified that can be added to Google Authenticator mobile app.
		So(dataURI, ShouldEqual, "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAGQAAABkEAAAAAAFGRbLAAAEIElEQVR4nORb7XLkMAiTb/L+r5ybnZ57BEuCbf90KH/qxJ+LbEA4ve4bI+QC1tLVr5+563eZvdvll+w2uxzrorhx2Xh+nX+aP/jHy/Xxx2ks1sd2qg9DgyGT28X3r3cMAb3OaYhAIJDLIFqPGlR72Wlf9Vdj8XVORERJRABCU8xixf4I2kdCMp6/7rk45Tch4uy50irTrjqDlR/ryURE3K9nKICgFK1QHrODUG7P1sTXOQ0Rt0+z1rPFcfWAt0xu3Dy/X+cYRFZtF5SnZfu/auPqVRvW7pQxiFx1jFPFPNni5H3P+qDhS/K5UDEeZvGRdd/slx7NDOtT/RXPUFKxSsc2Z52RLU7rivWxfe4iY/ac+7s+kP5mDCL/zghMnMOsGopoYEt1rjoZHBdR/287DZGOxiskvuLFWb9qLI7yRETQzAzCcPRuJqSbmVEoxnE+2k1CBEUetuLsaPqOijW+mxN71o9B5NLxkYtiFQOM7Zwv2s+Vx87zq/pZiDgLtOVO+dmOKIvDtOxkpfwyEqLDrJbINLL9XXlqCAuncla57JBhMdfznI1BxDDEDnOE8SGqzmVKnFfP483l7K1mxd0gCm+OL2RNUNwkP1c+BpGrpzWl3ahFta+zD8jzxP4wdy6q3zyr9VlsZjFYexirU0XLLt6qIu6ZWZTPYpmpeM8vKKkYZGXd2NyDELl4LFV5Z1bneDh7dr6hmvP0JZMQiaKYX352ViRrs4qkK3+h1gOMjLXSrW7ew7GsmB3TlNJu3NuMdbIogK0nzzHojJDot7LdMPtVWZ1K9N7X6KyZuV/zlamLOiEQYL4i91FzdOZXUfEsRPDG3Xqsh9jLlbd2UcNKGUhm2bi/mobIlipr4rw3GwPmNreKzfJ86vytcZlGlpVQFiPWb6nOierbRXeRrOQpkxBBkT9CcceneEWnL0Oq4vCx/UA/Iu7ZHfN7dC84iyqjcV+Z5/Z+bQwiVx1hOguEgrlli3Pfnpd3kOERwyREoij+7Dx3Luc2DvGqL0T2/SxPRQQNdgZj35WF63DxPGZnXffIvJZCoOOF2bNCx2k5P3fP1Bqb1wKxMCz2eScOguA3yms7L15nU8YgEr7XUhqpYic0bpoc22T987s8xtl+DCLkP0NzPOSQUMIihDjPd8/HuYZJiHS5OPPOKh5S/uEmX83FsvMVjsvPQsTFRyi8LBq3vVsyt6n4vfNf506YhIjnwvzZnYcsWcvZEnb6sXGec09CBI3IVPkR5Y2jsLGr2KnyHec80xCBye3eJKuo3uWxHN/oen02zikTEVHS5SlK2x1P3okoovyuLIoSl1WBsU5o5q7cOHn+tXLcNhGRShNMu0yzyppljqPG38JYo+JJ8xBxe59Fpnqvcs1mYegofq8Qe1rGMYg0v43/+fI3AAD//6d0Yx9AJENIAAAAAElFTkSuQmCC")
	})
}
