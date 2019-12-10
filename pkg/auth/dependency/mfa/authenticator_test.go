package mfa

import (
	"testing"
	"time"

	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMaskAuthenticators(t *testing.T) {
	date := time.Date(2019, 6, 1, 0, 0, 0, 0, time.UTC)
	Convey("MaskAuthenticators", t, func() {
		input := []Authenticator{
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
		expected := []Authenticator{
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

func TestCanAddAuthenticator(t *testing.T) {
	type Existing struct {
		TOTP     int
		OOBSMS   int
		OOBEmail int
	}
	type Limit struct {
		Total    int
		TOTP     int
		OOBSMS   int
		OOBEmail int
	}
	type Case struct {
		Enabled  bool
		Existing Existing
		New      Authenticator
		Limit    Limit
		Expected bool
	}

	f := func(c Case) {
		var authenticators []Authenticator
		for i := 0; i < c.Existing.TOTP; i++ {
			authenticators = append(authenticators, TOTPAuthenticator{})
		}
		for i := 0; i < c.Existing.OOBSMS; i++ {
			authenticators = append(authenticators, OOBAuthenticator{
				Channel: coreAuth.AuthenticatorOOBChannelSMS,
			})
		}
		for i := 0; i < c.Existing.OOBEmail; i++ {
			authenticators = append(authenticators, OOBAuthenticator{
				Channel: coreAuth.AuthenticatorOOBChannelEmail,
			})
		}

		newA := c.New

		mfaConfiguration := &config.MFAConfiguration{
			Enabled: c.Enabled,
			Maximum: &c.Limit.Total,
			TOTP: &config.MFATOTPConfiguration{
				Maximum: &c.Limit.TOTP,
			},
			OOB: &config.MFAOOBConfiguration{
				SMS: &config.MFAOOBSMSConfiguration{
					Maximum: &c.Limit.OOBSMS,
				},
				Email: &config.MFAOOBEmailConfiguration{
					Maximum: &c.Limit.OOBEmail,
				},
			},
		}

		actual := CanAddAuthenticator(authenticators, newA, mfaConfiguration)
		So(actual, ShouldEqual, c.Expected)
	}

	cases := []Case{
		Case{
			Enabled: true,
			Existing: Existing{
				TOTP:     0,
				OOBSMS:   0,
				OOBEmail: 0,
			},
			Limit: Limit{
				Total:    1,
				TOTP:     1,
				OOBSMS:   0,
				OOBEmail: 0,
			},
			New:      TOTPAuthenticator{},
			Expected: true,
		},

		Case{
			Enabled: false,
			Existing: Existing{
				TOTP:     0,
				OOBSMS:   0,
				OOBEmail: 0,
			},
			Limit: Limit{
				Total:    1,
				TOTP:     1,
				OOBSMS:   0,
				OOBEmail: 0,
			},
			New:      TOTPAuthenticator{},
			Expected: false,
		},

		Case{
			Enabled: true,
			Existing: Existing{
				TOTP:     1,
				OOBSMS:   0,
				OOBEmail: 0,
			},
			Limit: Limit{
				Total:    1,
				TOTP:     1,
				OOBSMS:   0,
				OOBEmail: 0,
			},
			New:      TOTPAuthenticator{},
			Expected: false,
		},

		Case{
			Enabled: true,
			Existing: Existing{
				TOTP:     0,
				OOBSMS:   0,
				OOBEmail: 1,
			},
			Limit: Limit{
				Total:    1,
				TOTP:     0,
				OOBSMS:   0,
				OOBEmail: 1,
			},
			New:      TOTPAuthenticator{},
			Expected: false,
		},

		Case{
			Enabled: true,
			Existing: Existing{
				TOTP:     0,
				OOBSMS:   0,
				OOBEmail: 1,
			},
			Limit: Limit{
				Total:    2,
				TOTP:     0,
				OOBSMS:   0,
				OOBEmail: 2,
			},
			New: OOBAuthenticator{
				Channel: coreAuth.AuthenticatorOOBChannelEmail,
			},
			Expected: true,
		},

		Case{
			Enabled: true,
			Existing: Existing{
				TOTP:     1,
				OOBSMS:   0,
				OOBEmail: 0,
			},
			Limit: Limit{
				Total:    2,
				TOTP:     1,
				OOBSMS:   1,
				OOBEmail: 0,
			},
			New: OOBAuthenticator{
				Channel: coreAuth.AuthenticatorOOBChannelSMS,
			},
			Expected: true,
		},

		Case{
			Enabled: false,
			Existing: Existing{
				TOTP:     1,
				OOBSMS:   0,
				OOBEmail: 0,
			},
			Limit: Limit{
				Total:    2,
				TOTP:     1,
				OOBSMS:   1,
				OOBEmail: 0,
			},
			New: OOBAuthenticator{
				Channel: coreAuth.AuthenticatorOOBChannelSMS,
			},
			Expected: false,
		},

		Case{
			Enabled: true,
			Existing: Existing{
				TOTP:     98,
				OOBSMS:   0,
				OOBEmail: 0,
			},
			Limit: Limit{
				Total:    99,
				TOTP:     99,
				OOBSMS:   99,
				OOBEmail: 99,
			},
			New:      TOTPAuthenticator{},
			Expected: true,
		},
		Case{
			Enabled: true,
			Existing: Existing{
				TOTP:     99,
				OOBSMS:   0,
				OOBEmail: 0,
			},
			Limit: Limit{
				Total:    99,
				TOTP:     99,
				OOBSMS:   99,
				OOBEmail: 99,
			},
			New:      TOTPAuthenticator{},
			Expected: false,
		},
	}
	Convey("CanAddAuthenticator", t, func() {
		for _, c := range cases {
			f(c)
		}
	})
}

func TestIsDeletingOnlyActivatedAuthenticator(t *testing.T) {
	type Case struct {
		Authenticators []Authenticator
		Authenticator  Authenticator
		Expected       bool
	}
	cases := []Case{
		Case{
			Authenticators: []Authenticator{
				TOTPAuthenticator{
					ID:        "totp",
					Activated: true,
				},
			},
			Authenticator: TOTPAuthenticator{
				ID:        "totp",
				Activated: false,
			},
			Expected: false,
		},
		Case{
			Authenticators: []Authenticator{
				TOTPAuthenticator{
					ID:        "totp",
					Activated: true,
				},
			},
			Authenticator: TOTPAuthenticator{
				ID:        "totp",
				Activated: true,
			},
			Expected: true,
		},
	}
	f := func(c Case) {
		actual := IsDeletingOnlyActivatedAuthenticator(c.Authenticators, c.Authenticator)
		So(actual, ShouldEqual, c.Expected)
	}
	Convey("IsDeletingActivatedAuthenticator", t, func() {
		for _, c := range cases {
			f(c)
		}
	})
}
