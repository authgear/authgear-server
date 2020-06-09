package hook

import (
	"context"

	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

func ProvideHookProvider(
	ctx context.Context,
	sqlb db.SQLBuilder,
	sqle db.SQLExecutor,
	tConfig *config.TenantConfiguration,
	txContext db.TxContext,
	timeProvider time.Provider,
	users UserProvider,
	loginIDProvider LoginIDProvider,
	loggerFactory logging.Factory,
) Provider {
	return NewProvider(
		ctx,
		NewStore(sqlb, sqle),
		txContext,
		timeProvider,
		users,
		NewDeliverer(
			tConfig,
			timeProvider,
			NewMutator(
				tConfig.AppConfig.UserVerification,
				loginIDProvider,
				users,
			),
		),
		loggerFactory,
	)
}

var DependencySet = wire.NewSet(
	ProvideHookProvider,
	wire.Bind(new(auth.HookProvider), new(Provider)),
)
