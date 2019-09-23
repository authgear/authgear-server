package authnsession

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/core/auth"
)

func TestAuthnSessionToken(t *testing.T) {
	Convey("AuthnSessionToken", t, func() {
		secret := "secret"
		claims := Claims{
			AuthnSession: auth.AuthnSession{
				ClientID:                "clientid",
				UserID:                  "user",
				PrincipalID:             "principal",
				RequiredSteps:           []auth.AuthnSessionStep{"identity", "mfa"},
				FinishedSteps:           []auth.AuthnSessionStep{"identity"},
				SessionCreateReason:     "reason",
				AuthenticatorID:         "authenticator",
				AuthenticatorType:       "totp",
				AuthenticatorOOBChannel: "sms",
			},
		}
		token, err := NewAuthnSessionToken(secret, claims)
		So(err, ShouldBeNil)
		expected, err := ParseAuthnSessionToken(secret, token)
		So(err, ShouldBeNil)
		So(&claims, ShouldResemble, expected)
	})
}
