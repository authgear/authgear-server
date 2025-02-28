package authflowv2

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

type SettingsIdentityService interface {
	GetWithUserID(ctx context.Context, userID string, identityID string) (*identity.Info, error)
	GetBySpecWithUserID(ctx context.Context, userID string, spec *identity.Spec) (*identity.Info, error)
}
