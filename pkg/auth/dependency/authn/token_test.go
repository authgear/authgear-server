package authn

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
)

func TestSessionToken(t *testing.T) {
	Convey("session token", t, func() {
		secret := "secret"
		claims := sessionToken{
			Session: Session{
				ClientID:            "clientid",
				RequiredSteps:       []SessionStep{"identity", "mfa"},
				FinishedSteps:       []SessionStep{"identity"},
				SessionCreateReason: "reason",
				Attrs: session.Attrs{
					UserID:                  "user",
					PrincipalID:             "principal",
					AuthenticatorID:         "authenticator",
					AuthenticatorType:       "totp",
					AuthenticatorOOBChannel: "sms",
				},
			},
		}
		token, err := encodeSessionToken(secret, claims)
		So(err, ShouldBeNil)
		expected, err := decodeSessionToken(secret, token)
		So(err, ShouldBeNil)
		So(&claims, ShouldResemble, expected)
	})
}
