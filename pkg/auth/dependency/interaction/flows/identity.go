package flows

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity"
	"github.com/skygeario/skygear-server/pkg/core/authn"
)

type IdentityProvider interface {
	GetByClaims(typ authn.IdentityType, claims map[string]interface{}) (string, *identity.Info, error)
}
