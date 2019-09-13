package oauth

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

type GetByProviderOptions struct {
	ProviderType   string
	ProviderKeys   map[string]interface{}
	ProviderUserID string
}

type GetByUserOptions struct {
	ProviderType string
	ProviderKeys map[string]interface{}
	UserID       string
}

type Provider interface {
	principal.Provider
	GetPrincipalByProvider(options GetByProviderOptions) (*Principal, error)

	GetPrincipalByUser(options GetByUserOptions) (*Principal, error)
	CreatePrincipal(principal *Principal) error
	UpdatePrincipal(principal *Principal) error
	DeletePrincipal(principal *Principal) error
}

func ProviderKeysFromProviderConfig(c config.OAuthProviderConfiguration) map[string]interface{} {
	m := map[string]interface{}{}
	if c.Tenant != "" {
		m["tenant"] = c.Tenant
	}
	return m
}
