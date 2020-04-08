package password

import (
	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/passwordhistory"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	coreTime "github.com/skygeario/skygear-server/pkg/core/time"
)

func ProvidePasswordProvider(
	sqlb db.SQLBuilder,
	sqle db.SQLExecutor,
	timeProvider coreTime.Provider,
	passwordHistoryStore passwordhistory.Store,
	loggerFactory logging.Factory,
	config *config.TenantConfiguration,
	reservedNameChecker *loginid.ReservedNameChecker,
) Provider {
	return NewProvider(
		timeProvider,
		NewStore(sqlb, sqle),
		passwordHistoryStore,
		loggerFactory,
		config.AppConfig.Identity.LoginID.Keys,
		config.AppConfig.Identity.LoginID.Types,
		config.AppConfig.Authenticator.Password.Policy.HistorySize > 0 ||
			config.AppConfig.Authenticator.Password.Policy.HistoryDays > 0,
		reservedNameChecker,
	)
}

var DependencySet = wire.NewSet(ProvidePasswordProvider)
