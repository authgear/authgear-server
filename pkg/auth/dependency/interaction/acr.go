package interaction

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/oidc"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func DeriveACR(primary *authenticator.Info, secondary *authenticator.Info) (acr string) {
	if secondary != nil {
		if secondary.Type != authn.AuthenticatorTypeBearerToken {
			acr = oidc.ACRMFA
		}
	}
	return
}
