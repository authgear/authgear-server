package authenticator

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/authn"
)

func TestAuthenticatorEqualTrue(t *testing.T) {
	Convey("AuthenticatorEqualTrue", t, func() {
		cases := []struct {
			A *Info
			B *Info
		}{
			// Password with the same primary/secondary tag.
			{
				&Info{
					Type: authn.AuthenticatorTypePassword,
					Tag: []string{
						TagPrimaryAuthenticator,
					},
				},
				&Info{
					Type: authn.AuthenticatorTypePassword,
					Tag: []string{
						TagPrimaryAuthenticator,
					},
				},
			},
			{
				&Info{
					Type: authn.AuthenticatorTypePassword,
					Tag: []string{
						TagSecondaryAuthenticator,
					},
				},
				&Info{
					Type: authn.AuthenticatorTypePassword,
					Tag: []string{
						TagSecondaryAuthenticator,
					},
				},
			},

			// TOTP with the same secret.
			{
				&Info{
					Type: authn.AuthenticatorTypeTOTP,
					Tag: []string{
						TagPrimaryAuthenticator,
					},
					Secret: "secret",
				},
				&Info{
					Type: authn.AuthenticatorTypeTOTP,
					Tag: []string{
						TagPrimaryAuthenticator,
					},
					Secret: "secret",
				},
			},
			{
				&Info{
					Type: authn.AuthenticatorTypeTOTP,
					Tag: []string{
						TagSecondaryAuthenticator,
					},
					Secret: "secret",
				},
				&Info{
					Type: authn.AuthenticatorTypeTOTP,
					Tag: []string{
						TagSecondaryAuthenticator,
					},
					Secret: "secret",
				},
			},

			// OOB with the same channel and target.
			{
				&Info{
					Type: authn.AuthenticatorTypeOOB,
					Tag: []string{
						TagPrimaryAuthenticator,
					},
					Props: map[string]interface{}{
						AuthenticatorPropOOBOTPChannelType: string(authn.AuthenticatorOOBChannelEmail),
						AuthenticatorPropOOBOTPEmail:       "user@example",
						AuthenticatorPropOOBOTPPhone:       "",
					},
				},
				&Info{
					Type: authn.AuthenticatorTypeOOB,
					Tag: []string{
						TagPrimaryAuthenticator,
					},
					Props: map[string]interface{}{
						AuthenticatorPropOOBOTPChannelType: string(authn.AuthenticatorOOBChannelEmail),
						AuthenticatorPropOOBOTPEmail:       "user@example",
						AuthenticatorPropOOBOTPPhone:       "",
					},
				},
			},

			{
				&Info{
					Type: authn.AuthenticatorTypeOOB,
					Tag: []string{
						TagPrimaryAuthenticator,
					},
					Props: map[string]interface{}{
						AuthenticatorPropOOBOTPChannelType: string(authn.AuthenticatorOOBChannelSMS),
						AuthenticatorPropOOBOTPEmail:       "",
						AuthenticatorPropOOBOTPPhone:       "+85299887766",
					},
				},
				&Info{
					Type: authn.AuthenticatorTypeOOB,
					Tag: []string{
						TagPrimaryAuthenticator,
					},
					Props: map[string]interface{}{
						AuthenticatorPropOOBOTPChannelType: string(authn.AuthenticatorOOBChannelSMS),
						AuthenticatorPropOOBOTPEmail:       "",
						AuthenticatorPropOOBOTPPhone:       "+85299887766",
					},
				},
			},

			{
				&Info{
					Type: authn.AuthenticatorTypeOOB,
					Tag: []string{
						TagSecondaryAuthenticator,
					},
					Props: map[string]interface{}{
						AuthenticatorPropOOBOTPChannelType: string(authn.AuthenticatorOOBChannelEmail),
						AuthenticatorPropOOBOTPEmail:       "user@example",
						AuthenticatorPropOOBOTPPhone:       "",
					},
				},
				&Info{
					Type: authn.AuthenticatorTypeOOB,
					Tag: []string{
						TagSecondaryAuthenticator,
					},
					Props: map[string]interface{}{
						AuthenticatorPropOOBOTPChannelType: string(authn.AuthenticatorOOBChannelEmail),
						AuthenticatorPropOOBOTPEmail:       "user@example",
						AuthenticatorPropOOBOTPPhone:       "",
					},
				},
			},

			{
				&Info{
					Type: authn.AuthenticatorTypeOOB,
					Tag: []string{
						TagSecondaryAuthenticator,
					},
					Props: map[string]interface{}{
						AuthenticatorPropOOBOTPChannelType: string(authn.AuthenticatorOOBChannelSMS),
						AuthenticatorPropOOBOTPEmail:       "",
						AuthenticatorPropOOBOTPPhone:       "+85299887766",
					},
				},
				&Info{
					Type: authn.AuthenticatorTypeOOB,
					Tag: []string{
						TagSecondaryAuthenticator,
					},
					Props: map[string]interface{}{
						AuthenticatorPropOOBOTPChannelType: string(authn.AuthenticatorOOBChannelSMS),
						AuthenticatorPropOOBOTPEmail:       "",
						AuthenticatorPropOOBOTPPhone:       "+85299887766",
					},
				},
			},
		}

		for _, c := range cases {
			So(c.A.Equal(c.B), ShouldBeTrue)
		}
	})
}

func TestAuthenticatorEqualFalse(t *testing.T) {
	Convey("AuthenticatorEqualFalse", t, func() {
		cases := []struct {
			A *Info
			B *Info
		}{
			// Different types.
			{
				&Info{
					Type: authn.AuthenticatorTypePassword,
				},
				&Info{
					Type: authn.AuthenticatorTypeTOTP,
				},
			},

			// Different primary/secondary tag.
			{
				&Info{
					Type: authn.AuthenticatorTypePassword,
					Tag: []string{
						TagPrimaryAuthenticator,
					},
				},
				&Info{
					Type: authn.AuthenticatorTypePassword,
					Tag: []string{
						TagSecondaryAuthenticator,
					},
				},
			},

			// TOTP with different secret.
			{
				&Info{
					Type: authn.AuthenticatorTypeTOTP,
					Tag: []string{
						TagPrimaryAuthenticator,
					},
					Secret: "secret1",
				},
				&Info{
					Type: authn.AuthenticatorTypeTOTP,
					Tag: []string{
						TagPrimaryAuthenticator,
					},
					Secret: "secret2",
				},
			},

			// OOB with the same channel but different target.
			{
				&Info{
					Type: authn.AuthenticatorTypeOOB,
					Tag: []string{
						TagPrimaryAuthenticator,
					},
					Props: map[string]interface{}{
						AuthenticatorPropOOBOTPChannelType: string(authn.AuthenticatorOOBChannelEmail),
						AuthenticatorPropOOBOTPEmail:       "user1@example",
						AuthenticatorPropOOBOTPPhone:       "",
					},
				},
				&Info{
					Type: authn.AuthenticatorTypeOOB,
					Tag: []string{
						TagPrimaryAuthenticator,
					},
					Props: map[string]interface{}{
						AuthenticatorPropOOBOTPChannelType: string(authn.AuthenticatorOOBChannelEmail),
						AuthenticatorPropOOBOTPEmail:       "user2@example",
						AuthenticatorPropOOBOTPPhone:       "",
					},
				},
			},
		}

		for _, c := range cases {
			So(c.A.Equal(c.B), ShouldBeFalse)
		}
	})
}
