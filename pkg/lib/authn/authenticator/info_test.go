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
					TOTP: &TOTP{
						Secret: "secret",
					},
				},
				&Info{
					Type: model.AuthenticatorTypeTOTP,
					Kind: KindPrimary,
					TOTP: &TOTP{
						Secret: "secret",
					},
				},
			},
			{
				&Info{
					Type: model.AuthenticatorTypeTOTP,
					Kind: KindSecondary,
					TOTP: &TOTP{
						Secret: "secret",
					},
				},
				&Info{
					Type: model.AuthenticatorTypeTOTP,
					Kind: KindSecondary,
					TOTP: &TOTP{
						Secret: "secret",
					},
				},
			},

			// OOB with the same channel and target.
			{
				&Info{
					Type: model.AuthenticatorTypeOOBEmail,
					Kind: KindPrimary,
					OOBOTP: &OOBOTP{
						Email: "user@example",
					},
				},
				&Info{
					Type: model.AuthenticatorTypeOOBEmail,
					Kind: KindPrimary,
					OOBOTP: &OOBOTP{
						Email: "user@example",
					},
				},
			},

			{
				&Info{
					Type: model.AuthenticatorTypeOOBSMS,
					Kind: KindPrimary,
					OOBOTP: &OOBOTP{
						Phone: "+85299887766",
					},
				},
				&Info{
					Type: model.AuthenticatorTypeOOBSMS,
					Kind: KindPrimary,
					OOBOTP: &OOBOTP{
						Phone: "+85299887766",
					},
				},
			},

			{
				&Info{
					Type: model.AuthenticatorTypeOOBEmail,
					Kind: KindSecondary,
					OOBOTP: &OOBOTP{
						Email: "user@example",
					},
				},
				&Info{
					Type: model.AuthenticatorTypeOOBEmail,
					Kind: KindSecondary,
					OOBOTP: &OOBOTP{
						Email: "user@example",
					},
				},
			},

			{
				&Info{
					Type: model.AuthenticatorTypeOOBSMS,
					Kind: KindSecondary,
					OOBOTP: &OOBOTP{
						Phone: "+85299887766",
					},
				},
				&Info{
					Type: model.AuthenticatorTypeOOBSMS,
					Kind: KindSecondary,
					OOBOTP: &OOBOTP{
						Phone: "+85299887766",
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
					TOTP: &TOTP{
						Secret: "secret1",
					},
				},
				&Info{
					Type: model.AuthenticatorTypeTOTP,
					Kind: KindPrimary,
					TOTP: &TOTP{
						Secret: "secret2",
					},
				},
			},

			// OOB with the same channel but different target.
			{
				&Info{
					Type: model.AuthenticatorTypeOOBEmail,
					Kind: KindPrimary,
					OOBOTP: &OOBOTP{
						Email: "user1@example",
					},
				},
				&Info{
					Type: model.AuthenticatorTypeOOBEmail,
					Kind: KindPrimary,
					OOBOTP: &OOBOTP{
						Email: "user2@example",
					},
				},
			},
		}

		for _, c := range cases {
			So(c.A.Equal(c.B), ShouldBeFalse)
		}
	})
}

func TestInfoToPublicClaimsPasskey(t *testing.T) {
	Convey("Info.ToPublicClaims of a passkey", t, func() {
		makeInfo := func(aaguid string) *Info {
			return (&Passkey{
				ID:           "id",
				UserID:       "userid",
				Kind:         string(KindPrimary),
				CredentialID: "credentialid",
				AAGUID:       aaguid,
			}).ToInfo()
		}

		Convey("has the AAGUID and the provider name of a known AAGUID", func() {
			claims := makeInfo("ea9b8d66-4d01-1d21-3ce4-b6b48cb575d4").ToPublicClaims()
			So(claims[AuthenticatorClaimPasskeyAAGUID], ShouldEqual, "ea9b8d66-4d01-1d21-3ce4-b6b48cb575d4")
			So(claims[AuthenticatorClaimPasskeyProviderName], ShouldEqual, "Google Password Manager")
		})

		Convey("has no provider name for the all-zero AAGUID", func() {
			claims := makeInfo("00000000-0000-0000-0000-000000000000").ToPublicClaims()
			So(claims[AuthenticatorClaimPasskeyAAGUID], ShouldEqual, "00000000-0000-0000-0000-000000000000")
			So(claims, ShouldNotContainKey, AuthenticatorClaimPasskeyProviderName)
		})

		Convey("has neither claim without an AAGUID", func() {
			claims := makeInfo("").ToPublicClaims()
			So(claims, ShouldNotContainKey, AuthenticatorClaimPasskeyAAGUID)
			So(claims, ShouldNotContainKey, AuthenticatorClaimPasskeyProviderName)
		})
	})
}
