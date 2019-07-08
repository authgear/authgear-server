package customtoken

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
)

const providerName string = "custom_token"

type Provider interface {
	principal.Provider
	Decode(tokenString string) (SSOCustomTokenClaims, error)
	CreatePrincipal(principal Principal) error
	GetPrincipalByTokenPrincipalID(tokenPrincipalID string) (*Principal, error)
}
