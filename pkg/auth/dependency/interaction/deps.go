package interaction

import (
	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/clock"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/logging"
)

func ProvideProvider(
	s Store,
	t clock.Clock,
	lf logging.Factory,
	ip IdentityProvider,
	ap AuthenticatorProvider,
	up UserProvider,
	oob OOBProvider,
	c *config.TenantConfiguration,
	hp HookProvider,
) *Provider {
	return &Provider{
		Store:         s,
		Clock:         t,
		Logger:        lf.NewLogger("interaction"),
		Identity:      ip,
		Authenticator: ap,
		User:          up,
		OOB:           oob,
		Hooks:         hp,
		Config:        c.AppConfig.Authentication,
	}
}

var DependencySet = wire.NewSet(
	ProvideProvider,
)
