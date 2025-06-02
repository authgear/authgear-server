package authenticationflow

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

// FIXME(dev-2754): Delete this interface
type IdentitySpecGetter interface {
	GetIdentitySpecs(ctx context.Context, deps *Dependencies, flows Flows) []*identity.Spec
}
