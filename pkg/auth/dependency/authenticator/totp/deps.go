package totp

import (
	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

func ProvideProvider(
	sqlb db.SQLBuilder,
	sqle db.SQLExecutor,
	t time.Provider,
	c *config.TenantConfiguration,
) *Provider {
	return &Provider{
		Store:  &Store{SQLBuilder: sqlb, SQLExecutor: sqle},
		Time:   t,
		Config: c.AppConfig.Authenticator.TOTP,
	}
}

var DependencySet = wire.NewSet(ProvideProvider)
