package flows

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

type IdentityProvider interface {
	GetByClaims(typ authn.IdentityType, claims map[string]interface{}) (string, *identity.Info, error)
}
