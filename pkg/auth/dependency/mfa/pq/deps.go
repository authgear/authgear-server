package pq

import (
	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/mfa"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

func ProvideStore(
	config *config.TenantConfiguration,
	sqlb db.SQLBuilder,
	sqle db.SQLExecutor,
	tp time.Provider,
) mfa.Store {
	return NewStore(config.AppConfig.MFA, sqlb, sqle, tp)
}

var DependencySet = wire.NewSet(ProvideStore)
