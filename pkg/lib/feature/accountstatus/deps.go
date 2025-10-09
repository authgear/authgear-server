package accountstatus

import (
	"context"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/util/backgroundjob"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

func NewRunner(ctx context.Context, runnableFactory backgroundjob.RunnableFactory) *backgroundjob.Runner {
	return backgroundjob.NewRunner(
		ctx,
		runnableFactory,
	)
}

func NewRunnableFactory(
	pool *db.Pool,
	globalDBCredentials *config.GlobalDatabaseCredentialsEnvironmentConfig,
	databaseCfg *config.DatabaseEnvironmentConfig,
	clock clock.Clock,
	appContextResolver AppContextResolver,
	userServiceFactory UserServiceFactory,
) backgroundjob.RunnableFactory {
	factory := func() backgroundjob.Runnable {
		return newRunnable(pool, globalDBCredentials, databaseCfg, clock, appContextResolver, userServiceFactory)
	}
	return factory
}

var DependencySet = wire.NewSet(
	NewRunnableFactory,
	NewRunner,
)

var RunnableDependencySet = wire.NewSet(
	globaldb.DependencySet,
	wire.Struct(new(Store), "*"),
	wire.Struct(new(Runnable), "*"),
	wire.Bind(new(backgroundjob.Runnable), new(*Runnable)),
)
