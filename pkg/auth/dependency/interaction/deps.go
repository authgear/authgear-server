package interaction

import (
	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

func ProvideProvider(
	s Store,
	t time.Provider,
	ip IdentityProvider,
	ap AuthenticatorProvider,
	c *config.TenantConfiguration,
) *Provider {
	return &Provider{
		Store:         s,
		Time:          t,
		Identity:      ip,
		Authenticator: ap,
		Config:        c.AppConfig.Authentication,
	}
}

var DependencySet = wire.NewSet(
	ProvideProvider,
)
