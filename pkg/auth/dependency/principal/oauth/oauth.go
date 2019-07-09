package oauth

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
)

const providerName string = "oauth"

type Provider interface {
	principal.Provider
	GetPrincipalByProviderUserID(providerName string, providerUserID string) (*Principal, error)
	GetPrincipalByUserID(providerName string, userID string) (*Principal, error)
	CreatePrincipal(principal Principal) error
	UpdatePrincipal(principal *Principal) error
	DeletePrincipal(principal *Principal) error
}
