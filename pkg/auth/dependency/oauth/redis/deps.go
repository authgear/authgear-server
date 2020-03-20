package redis

import (
	"context"

	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	coretime "github.com/skygeario/skygear-server/pkg/core/time"
)

func ProvideGrantStore(
	ctx context.Context,
	cfg *config.TenantConfiguration,
	sqlb db.SQLBuilder,
	sqle db.SQLExecutor,
	t coretime.Provider,
) *GrantStore {
	return &GrantStore{
		Context:     ctx,
		AppID:       cfg.AppID,
		SQLBuilder:  sqlb,
		SQLExecutor: sqle,
		Time:        t,
	}
}

var DependencySet = wire.NewSet(
	ProvideGrantStore,
	wire.Bind(new(oauth.CodeGrantStore), new(*GrantStore)),
)
