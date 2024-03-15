package secretcode

import (
	"encoding/base32"
	"errors"
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
		totp, _ := NewTOTPFromSecret(fixtureSecret)

		Convey("NewTOTPSecretFromRNG", func() {
			totp, err := NewTOTPFromRNG()
			So(err, ShouldBeNil)
			So(totp.Secret, ShouldNotBeEmpty)
			// The secret is of 160 bits
			// Base32 groups 5 bits into 1 character.
			// So the length should be 160/5 = 32.
			So(len(totp.Secret), ShouldEqual, 32)
		})

		Convey("NewTOTPFromSecret", func() {
			_, err := NewTOTPFromSecret("!")
			var corruptInputError base32.CorruptInputError
			So(errors.As(err, &corruptInputError), ShouldBeTrue)
			So(err, ShouldBeError, "illegal base32 data at input byte 0")
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

func TestTOTPGetURI(t *testing.T) {
	Convey("GetURI", t, func() {
		// Use a fixed secret to make the test stable.
		// This must be at least 20 bytes.
		rawSecret := "01234567890123456789"
		enc := base32.StdEncoding.WithPadding(base32.NoPadding)
		secret := enc.EncodeToString([]byte(rawSecret))
		totp, _ := NewTOTPFromSecret(secret)

		test := func(opts URIOptions, expected string) {
			u := totp.GetURI(opts)
			So(u.String(), ShouldEqual, expected)
		}

		test(URIOptions{
			Issuer:      "test",
			AccountName: "john.doe@example.com",
		}, "otpauth://totp/john.doe@example.com?algorithm=SHA1&digits=6&issuer=test&period=30&secret=GAYTEMZUGU3DOOBZGAYTEMZUGU3DOOBZ")

		test(URIOptions{
			Issuer:      "http://localhost:3100",
			AccountName: "john.doe@example.com",
		}, "otpauth://totp/john.doe@example.com?algorithm=SHA1&digits=6&issuer=http%3A%2F%2Flocalhost%3A3100&period=30&secret=GAYTEMZUGU3DOOBZGAYTEMZUGU3DOOBZ")
	})
}

func TestTOTPQRCodeImage(t *testing.T) {
	Convey("QRCodeImage", t, func() {
		// Use a fixed secret to make the test stable.
		// This must be at least 20 bytes.
		rawSecret := "01234567890123456789"
		enc := base32.StdEncoding.WithPadding(base32.NoPadding)
		secret := enc.EncodeToString([]byte(rawSecret))
		totp, _ := NewTOTPFromSecret(secret)

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
		So(dataURI, ShouldEqual, "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAGQAAABkEAAAAAAFGRbLAAADpklEQVR4nOxbW27jMAxkFr3/lbNYGEYUah50sUCDcecnrUVJNofiy8nX81kR+PPTN/C/8Psgn4aYB/l6/fl4cLF/LqGPn9fYJ5NzON3PZM7LVSUyUm9P+ALSiHLap3zX5npdzamFjfl9pTJStds3wqmxc5xpD52Pzti6Rmdifl/JjDAgja3/O41X7XPRJ1rb446MdO0j7Xb5PrayxeLNJNbsSGaE2SWL7n3csYbGVYxx93UglRFlnyqvQp99XhnPp9bx5yaREeezXZbL1pn8z6L/5L4OxDDyeD3v1CZV/JjWHGrNAuemX9vnJTJSRMM1yHtUTFCaRPHDMYqz5BhGSM3ePYiy4S6P1lvX7ePoXE5yuvexGEbAGZlG3TJeqsv19ZSs69is1w/EMAK6KKwn1cdZxJ7EFq9h3jnB3jOVEWSnyINMzoaKSX0N5CHReSyah8UwYiJ7h9IqklVjaD/Fhq4iExlxGafyMMqOV3nmEV19wxgKPCOiQnTRnNk/ixmua4K03ffhjCcygqBY6XJdvo/hOoJnuG6vdyQy0jXT7VHVG6p+Qd6whKebyO4eLpERFENq0BlhWapbl8m7qhDvG8PIY9fnJMrW4O2sqtlZjoXWmu0fw4j55kO3VWavXXYyHwF5JzR/3yuRkRXMi6hMlXkmV/M7Zlmdnp/9lqkxStQsJXIpNL+Pu6yhy+7rxDDSKsQTLpqX0fRkLsvJ1BpcPoYRkGt1oL4WslN1ftAeqnc2zR5eiGEE5Fol6ud+Dc1ZoeIO3lnc6j36WuYbdCW+Y3KFIbcmmsf6YDjGJDPiek3MtpkNO4/kYk7dq/cr3uqufzOtsQ4kij0lWHGZsIo9BxIZOaE80qTDweawqM7m90xBdWaCGCFvrBwrJyYeSXknVg2ivA3dXyAjpvfbbbUudB6VbTtc6QkciGFEZL+uilN1+4rvsODO1T1q9m3IfPNNZbBoPsuOVaxy7GZ7rbpgkyy6s7hQw3eCrO5BrIeeERJHkJacdh/Dd40rlEdEn/xExzAivNYmanrBK1QMUnnbd2qZAzGMXPitboE48rz43qTLTuSR19rtKJGRGrxdRZ0UVn+jNft87410RZof2cv4fSXjvJIC2/OG2e+FX4YW+XYO80p9Hrfvdwa7R2P51n3OCIP35dfXUhH+BGJlZz2ZEaVplVuV8GY16C72ayU6OTtSGVF+X/W3XBy4kjmj+3H9gSBGLtQjn40YRn4f5NMQ8yB/AwAA//8zLpMhZCHq2gAAAABJRU5ErkJggg==")
	})
}
