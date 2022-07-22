package authenticator

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
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
					Type: model.AuthenticatorTypePassword,
					Kind: KindPrimary,
				},
				&Info{
					Type: model.AuthenticatorTypePassword,
					Kind: KindPrimary,
				},
			},
			{
				&Info{
					Type: model.AuthenticatorTypePassword,
					Kind: KindSecondary,
				},
				&Info{
					Type: model.AuthenticatorTypePassword,
					Kind: KindSecondary,
				},
			},

			// TOTP with the same secret.
			{
				&Info{
					Type: model.AuthenticatorTypeTOTP,
					Kind: KindPrimary,
					Claims: map[ClaimKey]interface{}{
						AuthenticatorClaimTOTPSecret: "secret",
					},
				},
				&Info{
					Type: model.AuthenticatorTypeTOTP,
					Kind: KindPrimary,
					Claims: map[ClaimKey]interface{}{
						AuthenticatorClaimTOTPSecret: "secret",
					},
				},
			},
			{
				&Info{
					Type: model.AuthenticatorTypeTOTP,
					Kind: KindSecondary,
					Claims: map[ClaimKey]interface{}{
						AuthenticatorClaimTOTPSecret: "secret",
					},
				},
				&Info{
					Type: model.AuthenticatorTypeTOTP,
					Kind: KindSecondary,
					Claims: map[ClaimKey]interface{}{
						AuthenticatorClaimTOTPSecret: "secret",
					},
				},
			},

			// OOB with the same channel and target.
			{
				&Info{
					Type: model.AuthenticatorTypeOOBEmail,
					Kind: KindPrimary,
					Claims: map[ClaimKey]interface{}{
						AuthenticatorClaimOOBOTPEmail: "user@example",
						AuthenticatorClaimOOBOTPPhone: "",
					},
				},
				&Info{
					Type: model.AuthenticatorTypeOOBEmail,
					Kind: KindPrimary,
					Claims: map[ClaimKey]interface{}{
						AuthenticatorClaimOOBOTPEmail: "user@example",
						AuthenticatorClaimOOBOTPPhone: "",
					},
				},
			},

			{
				&Info{
					Type: model.AuthenticatorTypeOOBSMS,
					Kind: KindPrimary,
					Claims: map[ClaimKey]interface{}{
						AuthenticatorClaimOOBOTPEmail: "",
						AuthenticatorClaimOOBOTPPhone: "+85299887766",
					},
				},
				&Info{
					Type: model.AuthenticatorTypeOOBSMS,
					Kind: KindPrimary,
					Claims: map[ClaimKey]interface{}{
						AuthenticatorClaimOOBOTPEmail: "",
						AuthenticatorClaimOOBOTPPhone: "+85299887766",
					},
				},
			},

			{
				&Info{
					Type: model.AuthenticatorTypeOOBEmail,
					Kind: KindSecondary,
					Claims: map[ClaimKey]interface{}{
						AuthenticatorClaimOOBOTPEmail: "user@example",
						AuthenticatorClaimOOBOTPPhone: "",
					},
				},
				&Info{
					Type: model.AuthenticatorTypeOOBEmail,
					Kind: KindSecondary,
					Claims: map[ClaimKey]interface{}{
						AuthenticatorClaimOOBOTPEmail: "user@example",
						AuthenticatorClaimOOBOTPPhone: "",
					},
				},
			},

			{
				&Info{
					Type: model.AuthenticatorTypeOOBSMS,
					Kind: KindSecondary,
					Claims: map[ClaimKey]interface{}{
						AuthenticatorClaimOOBOTPEmail: "",
						AuthenticatorClaimOOBOTPPhone: "+85299887766",
					},
				},
				&Info{
					Type: model.AuthenticatorTypeOOBSMS,
					Kind: KindSecondary,
					Claims: map[ClaimKey]interface{}{
						AuthenticatorClaimOOBOTPEmail: "",
						AuthenticatorClaimOOBOTPPhone: "+85299887766",
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
					Type: model.AuthenticatorTypePassword,
				},
				&Info{
					Type: model.AuthenticatorTypeTOTP,
				},
			},

			// Different primary/secondary tag.
			{
				&Info{
					Type: model.AuthenticatorTypePassword,
					Kind: KindPrimary,
				},
				&Info{
					Type: model.AuthenticatorTypePassword,
					Kind: KindSecondary,
				},
			},

			// TOTP with different secret.
			{
				&Info{
					Type: model.AuthenticatorTypeTOTP,
					Kind: KindPrimary,
					Claims: map[ClaimKey]interface{}{
						AuthenticatorClaimTOTPSecret: "secret1",
					},
				},
				&Info{
					Type: model.AuthenticatorTypeTOTP,
					Kind: KindPrimary,
					Claims: map[ClaimKey]interface{}{
						AuthenticatorClaimTOTPSecret: "secret2",
					},
				},
			},

			// OOB with the same channel but different target.
			{
				&Info{
					Type: model.AuthenticatorTypeOOBEmail,
					Kind: KindPrimary,
					Claims: map[ClaimKey]interface{}{
						AuthenticatorClaimOOBOTPEmail: "user1@example",
						AuthenticatorClaimOOBOTPPhone: "",
					},
				},
				&Info{
					Type: model.AuthenticatorTypeOOBEmail,
					Kind: KindPrimary,
					Claims: map[ClaimKey]interface{}{
						AuthenticatorClaimOOBOTPEmail: "user2@example",
						AuthenticatorClaimOOBOTPPhone: "",
					},
				},
			},
		}

		for _, c := range cases {
			So(c.A.Equal(c.B), ShouldBeFalse)
		}
	})
}
