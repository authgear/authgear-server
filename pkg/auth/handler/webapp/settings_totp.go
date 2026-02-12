package webapp

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
)

type SettingsTOTPViewModel struct {
	Authenticators []*authenticator.Info
}
