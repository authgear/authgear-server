package interaction_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/interaction"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func TestDeriveAMR(t *testing.T) {
	Convey("DeriveAMR", t, func() {
		test := func(primary *authenticator.Info, secondary *authenticator.Info, expected []string) {
			actual := interaction.DeriveAMR(primary, secondary)
			So(actual, ShouldResemble, expected)
		}

		// OAuth
		test(nil, nil, []string{})

		// password
		test(&authenticator.Info{
			Type: authn.AuthenticatorTypePassword,
		}, nil, []string{"pwd"})

		// OOB
		test(&authenticator.Info{
			Type: authn.AuthenticatorTypeOOB,
			Props: map[string]interface{}{
				authenticator.AuthenticatorPropOOBOTPChannelType: string(authn.AuthenticatorOOBChannelEmail),
			},
		}, nil, []string{"otp"})

		// OOB
		test(&authenticator.Info{
			Type: authn.AuthenticatorTypeOOB,
			Props: map[string]interface{}{
				authenticator.AuthenticatorPropOOBOTPChannelType: string(authn.AuthenticatorOOBChannelSMS),
			},
		}, nil, []string{"otp", "sms"})

		// OOB + OOB
		test(&authenticator.Info{
			Type: authn.AuthenticatorTypeOOB,
			Props: map[string]interface{}{
				authenticator.AuthenticatorPropOOBOTPChannelType: string(authn.AuthenticatorOOBChannelSMS),
			},
		}, &authenticator.Info{
			Type: authn.AuthenticatorTypeOOB,
			Props: map[string]interface{}{
				authenticator.AuthenticatorPropOOBOTPChannelType: string(authn.AuthenticatorOOBChannelEmail),
			},
		}, []string{"mfa", "otp", "sms"})

		// password + TOTP
		test(&authenticator.Info{
			Type: authn.AuthenticatorTypePassword,
		}, &authenticator.Info{
			Type: authn.AuthenticatorTypeTOTP,
		}, []string{"mfa", "otp", "pwd"})

		// OAuth + bearer token
		test(nil, &authenticator.Info{
			Type: authn.AuthenticatorTypeBearerToken,
		}, []string{})
	})
}
