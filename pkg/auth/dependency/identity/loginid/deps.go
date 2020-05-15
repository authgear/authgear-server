package loginid

import (
	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/loginid"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

func ProvideProvider(
	sqlb db.SQLBuilder,
	sqle db.SQLExecutor,
	t time.Provider,
	c *config.TenantConfiguration,
	reservedNameChecker *loginid.ReservedNameChecker,
) *Provider {
	config := *c.AppConfig.Identity.LoginID
	return &Provider{
		Store:  &Store{SQLBuilder: sqlb, SQLExecutor: sqle},
		Time:   t,
		Config: config,
		LoginIDChecker: loginid.NewDefaultLoginIDChecker(
			config.Keys,
			config.Types,
			reservedNameChecker,
		),
		LoginIDNormalizerFactory: loginid.NewLoginIDNormalizerFactory(config.Keys, config.Types),
	}
}

var DependencySet = wire.NewSet(ProvideProvider)
