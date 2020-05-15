package password

import (
	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/passwordhistory"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

func ProvideProvider(
	sqlb db.SQLBuilder,
	sqle db.SQLExecutor,
	t time.Provider,
	lf logging.Factory,
	ph passwordhistory.Store,
	pc *audit.PasswordChecker,
	c *config.TenantConfiguration,
) *Provider {
	return &Provider{
		Store:           &Store{SQLBuilder: sqlb, SQLExecutor: sqle},
		Time:            t,
		Config:          c.AppConfig.Authenticator.Password,
		Logger:          lf.NewLogger("authenticator-password"),
		PasswordHistory: ph,
		PasswordChecker: pc,
	}
}

var DependencySet = wire.NewSet(ProvideProvider)
