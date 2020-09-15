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
					Kind: KindPrimary,
				},
				&Info{
					Type: authn.AuthenticatorTypePassword,
					Kind: KindPrimary,
				},
			},
			{
				&Info{
					Type: authn.AuthenticatorTypePassword,
					Kind: KindSecondary,
				},
				&Info{
					Type: authn.AuthenticatorTypePassword,
					Kind: KindSecondary,
				},
			},

			// TOTP with the same secret.
			{
				&Info{
					Type:   authn.AuthenticatorTypeTOTP,
					Kind:   KindPrimary,
					Secret: "secret",
				},
				&Info{
					Type:   authn.AuthenticatorTypeTOTP,
					Kind:   KindPrimary,
					Secret: "secret",
				},
			},
			{
				&Info{
					Type:   authn.AuthenticatorTypeTOTP,
					Kind:   KindSecondary,
					Secret: "secret",
				},
				&Info{
					Type:   authn.AuthenticatorTypeTOTP,
					Kind:   KindSecondary,
					Secret: "secret",
				},
			},

			// OOB with the same channel and target.
			{
				&Info{
					Type: authn.AuthenticatorTypeOOB,
					Kind: KindPrimary,
					Claims: map[string]interface{}{
						AuthenticatorClaimOOBOTPChannelType: string(authn.AuthenticatorOOBChannelEmail),
						AuthenticatorClaimOOBOTPEmail:       "user@example",
						AuthenticatorClaimOOBOTPPhone:       "",
					},
				},
				&Info{
					Type: authn.AuthenticatorTypeOOB,
					Kind: KindPrimary,
					Claims: map[string]interface{}{
						AuthenticatorClaimOOBOTPChannelType: string(authn.AuthenticatorOOBChannelEmail),
						AuthenticatorClaimOOBOTPEmail:       "user@example",
						AuthenticatorClaimOOBOTPPhone:       "",
					},
				},
			},

			{
				&Info{
					Type: authn.AuthenticatorTypeOOB,
					Kind: KindPrimary,
					Claims: map[string]interface{}{
						AuthenticatorClaimOOBOTPChannelType: string(authn.AuthenticatorOOBChannelSMS),
						AuthenticatorClaimOOBOTPEmail:       "",
						AuthenticatorClaimOOBOTPPhone:       "+85299887766",
					},
				},
				&Info{
					Type: authn.AuthenticatorTypeOOB,
					Kind: KindPrimary,
					Claims: map[string]interface{}{
						AuthenticatorClaimOOBOTPChannelType: string(authn.AuthenticatorOOBChannelSMS),
						AuthenticatorClaimOOBOTPEmail:       "",
						AuthenticatorClaimOOBOTPPhone:       "+85299887766",
					},
				},
			},

			{
				&Info{
					Type: authn.AuthenticatorTypeOOB,
					Kind: KindSecondary,
					Claims: map[string]interface{}{
						AuthenticatorClaimOOBOTPChannelType: string(authn.AuthenticatorOOBChannelEmail),
						AuthenticatorClaimOOBOTPEmail:       "user@example",
						AuthenticatorClaimOOBOTPPhone:       "",
					},
				},
				&Info{
					Type: authn.AuthenticatorTypeOOB,
					Kind: KindSecondary,
					Claims: map[string]interface{}{
						AuthenticatorClaimOOBOTPChannelType: string(authn.AuthenticatorOOBChannelEmail),
						AuthenticatorClaimOOBOTPEmail:       "user@example",
						AuthenticatorClaimOOBOTPPhone:       "",
					},
				},
			},

			{
				&Info{
					Type: authn.AuthenticatorTypeOOB,
					Kind: KindSecondary,
					Claims: map[string]interface{}{
						AuthenticatorClaimOOBOTPChannelType: string(authn.AuthenticatorOOBChannelSMS),
						AuthenticatorClaimOOBOTPEmail:       "",
						AuthenticatorClaimOOBOTPPhone:       "+85299887766",
					},
				},
				&Info{
					Type: authn.AuthenticatorTypeOOB,
					Kind: KindSecondary,
					Claims: map[string]interface{}{
						AuthenticatorClaimOOBOTPChannelType: string(authn.AuthenticatorOOBChannelSMS),
						AuthenticatorClaimOOBOTPEmail:       "",
						AuthenticatorClaimOOBOTPPhone:       "+85299887766",
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
					Kind: KindPrimary,
				},
				&Info{
					Type: authn.AuthenticatorTypePassword,
					Kind: KindSecondary,
				},
			},

			// TOTP with different secret.
			{
				&Info{
					Type:   authn.AuthenticatorTypeTOTP,
					Kind:   KindPrimary,
					Secret: "secret1",
				},
				&Info{
					Type:   authn.AuthenticatorTypeTOTP,
					Kind:   KindPrimary,
					Secret: "secret2",
				},
			},

			// OOB with the same channel but different target.
			{
				&Info{
					Type: authn.AuthenticatorTypeOOB,
					Kind: KindPrimary,
					Claims: map[string]interface{}{
						AuthenticatorClaimOOBOTPChannelType: string(authn.AuthenticatorOOBChannelEmail),
						AuthenticatorClaimOOBOTPEmail:       "user1@example",
						AuthenticatorClaimOOBOTPPhone:       "",
					},
				},
				&Info{
					Type: authn.AuthenticatorTypeOOB,
					Kind: KindPrimary,
					Claims: map[string]interface{}{
						AuthenticatorClaimOOBOTPChannelType: string(authn.AuthenticatorOOBChannelEmail),
						AuthenticatorClaimOOBOTPEmail:       "user2@example",
						AuthenticatorClaimOOBOTPPhone:       "",
					},
				},
			},
		}

		for _, c := range cases {
			So(c.A.Equal(c.B), ShouldBeFalse)
		}
	})
}
