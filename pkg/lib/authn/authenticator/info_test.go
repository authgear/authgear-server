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
