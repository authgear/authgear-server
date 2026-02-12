package webapp

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
)

type SettingsOOBOTPViewModel struct {
	OOBAuthenticatorType model.AuthenticatorType
	Authenticators       []*authenticator.Info
}
