package mfa

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestKeyURI(t *testing.T) {
	Convey("IsGoogleAuthenticatorCompatible", t, func() {
		keyURI := KeyURI{
			Type:        KeyURITypeTOTP,
			Issuer:      "",
			AccountName: "",
			Secret:      "JBSWY3DPEHPK3PXP",
			Algorithm:   KeyURIAlgorithmSHA1,
			Digits:      6,
			Counter:     "",
			Period:      30,
		}
		So(keyURI.IsGoogleAuthenticatorCompatible(), ShouldBeTrue)
	})
	Convey("ParseKeyURI", t, func() {
		Convey("minimal totp", func() {
			actual, err := ParseKeyURI("otpauth://totp/?secret=JBSWY3DPEHPK3PXP")
			So(err, ShouldBeNil)
			expected := KeyURI{
				Type:        KeyURITypeTOTP,
				Issuer:      "",
				AccountName: "",
				Secret:      "JBSWY3DPEHPK3PXP",
				Algorithm:   KeyURIAlgorithmSHA1,
				Digits:      6,
				Counter:     "",
				Period:      30,
			}
			So(*actual, ShouldResemble, expected)
		})
		Convey("minimal hotp", func() {
			actual, err := ParseKeyURI("otpauth://hotp/?secret=JBSWY3DPEHPK3PXP&counter=c")
			So(err, ShouldBeNil)
			expected := KeyURI{
				Type:        KeyURITypeHOTP,
				Issuer:      "",
				AccountName: "",
				Secret:      "JBSWY3DPEHPK3PXP",
				Algorithm:   KeyURIAlgorithmSHA1,
				Digits:      6,
				Counter:     "c",
				Period:      30,
			}
			So(*actual, ShouldResemble, expected)
		})
		Convey("full", func() {
			actual, err := ParseKeyURI("otpauth://totp/Example:user@example.com?secret=JBSWY3DPEHPK3PXP&algorithm=SHA512&digits=8&period=90")
			So(err, ShouldBeNil)
			expected := KeyURI{
				Type:        KeyURITypeTOTP,
				Issuer:      "Example",
				AccountName: "user@example.com",
				Secret:      "JBSWY3DPEHPK3PXP",
				Algorithm:   KeyURIAlgorithmSHA512,
				Digits:      8,
				Counter:     "",
				Period:      90,
			}
			So(*actual, ShouldResemble, expected)
		})
	})
	Convey("String", t, func() {
		i := "otpauth://totp/?secret=JBSWY3DPEHPK3PXP"
		u, err := ParseKeyURI(i)
		So(err, ShouldBeNil)
		So(u.String(), ShouldEqual, i)

		i = "otpauth://totp/%2Fjohndoe%2F?secret=JBSWY3DPEHPK3PXP"
		u, err = ParseKeyURI(i)
		So(err, ShouldBeNil)
		So(u.String(), ShouldEqual, i)

		i = "otpauth://totp/service:%2Fjohndoe%2F?secret=JBSWY3DPEHPK3PXP&issuer=service"
		u, err = ParseKeyURI(i)
		So(err, ShouldBeNil)
		So(u.String(), ShouldEqual, i)
	})
}
