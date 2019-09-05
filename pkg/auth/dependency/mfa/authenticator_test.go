package mfa

import (
	"testing"
	"time"

	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMaskAuthenticators(t *testing.T) {
	date := time.Date(2019, 6, 1, 0, 0, 0, 0, time.UTC)
	Convey("MaskAuthenticators", t, func() {
		input := []interface{}{
			TOTPAuthenticator{
				ID:          "totp",
				Type:        coreAuth.AuthenticatorTypeTOTP,
				CreatedAt:   date,
				ActivatedAt: &date,
				DisplayName: "totp",
			},
			OOBAuthenticator{
				ID:          "oobsms",
				Type:        coreAuth.AuthenticatorTypeOOB,
				CreatedAt:   date,
				ActivatedAt: &date,
				Channel:     coreAuth.AuthenticatorOOBChannelSMS,
				Phone:       "+85298765432",
			},
			OOBAuthenticator{
				ID:          "oobemail",
				Type:        coreAuth.AuthenticatorTypeOOB,
				CreatedAt:   date,
				ActivatedAt: &date,
				Channel:     coreAuth.AuthenticatorOOBChannelEmail,
				Email:       "johndoe@example.com",
			},
		}
		actual := MaskAuthenticators(input)
		expected := []interface{}{
			MaskedTOTPAuthenticator{
				ID:          "totp",
				Type:        coreAuth.AuthenticatorTypeTOTP,
				CreatedAt:   date,
				ActivatedAt: &date,
				DisplayName: "totp",
			},
			MaskedOOBAuthenticator{
				ID:          "oobsms",
				Type:        coreAuth.AuthenticatorTypeOOB,
				CreatedAt:   date,
				ActivatedAt: &date,
				Channel:     coreAuth.AuthenticatorOOBChannelSMS,
				MaskedPhone: "+8529876****",
			},
			MaskedOOBAuthenticator{
				ID:          "oobemail",
				Type:        coreAuth.AuthenticatorTypeOOB,
				CreatedAt:   date,
				ActivatedAt: &date,
				Channel:     coreAuth.AuthenticatorOOBChannelEmail,
				MaskedEmail: "joh****@example.com",
			},
		}
		So(actual, ShouldResemble, expected)
	})
}
