package workflow

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

type IdentityService interface {
	SearchBySpec(spec *identity.Spec) (exactMatch *identity.Info, otherMatches []*identity.Info, err error)
}

type Dependencies struct {
	Identities IdentityService
}
