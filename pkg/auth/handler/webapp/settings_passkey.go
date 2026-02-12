package webapp

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

type SettingsPasskeyViewModel struct {
	PasskeyIdentities []*identity.Info
}
