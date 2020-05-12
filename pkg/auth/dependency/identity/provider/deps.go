package provider

import (
	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

func ProvideProvider(
	c *config.TenantConfiguration,
	loginID LoginIDIdentityProvider,
	oauth OAuthIdentityProvider,
	anonymous AnonymousIdentityProvider,
) *Provider {
	return &Provider{
		Authentication: c.AppConfig.Authentication,
		Identity:       c.AppConfig.Identity,
		LoginID:        loginID,
		OAuth:          oauth,
		Anonymous:      anonymous,
	}
}

var DependencySet = wire.NewSet(
	ProvideProvider,
)
