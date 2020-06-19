package bearertoken

import (
	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/clock"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

func ProvideProvider(
	sqlb db.SQLBuilder,
	sqle db.SQLExecutor,
	t clock.Clock,
	c *config.TenantConfiguration,
) *Provider {
	return &Provider{
		Store:  &Store{SQLBuilder: sqlb, SQLExecutor: sqle},
		Clock:  t,
		Config: c.AppConfig.Authenticator.BearerToken,
	}
}

var DependencySet = wire.NewSet(ProvideProvider)
