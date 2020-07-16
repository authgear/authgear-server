package interaction

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

type Result struct {
	*authn.Attrs
	Identity               *identity.Info
	PrimaryAuthenticator   *authenticator.Info
	SecondaryAuthenticator *authenticator.Info
}
