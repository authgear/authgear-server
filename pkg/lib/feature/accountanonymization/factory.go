package accountanonymization

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/util/backgroundjob"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/log"
)

func NewRunnableFactory(
	pool *db.Pool,
	globalDBCredentials *config.GlobalDatabaseCredentialsEnvironmentConfig,
	databaseCfg *config.DatabaseEnvironmentConfig,
	logFactory *log.Factory,
	clock clock.Clock,
	appContextResolver AppContextResolver,
	userServiceFactory UserServiceFactory,
) backgroundjob.RunnableFactory {
	factory := func(ctx context.Context) backgroundjob.Runnable {
		runnableLogger := NewRunnableLogger(logFactory)
		handle := globaldb.NewHandle(
			ctx,
			pool,
			globalDBCredentials,
			databaseCfg,
			logFactory)
		sqlBuilder := globaldb.NewSQLBuilder(globalDBCredentials)
		sqlExecutor := globaldb.NewSQLExecutor(ctx, handle)
		store := &Store{
			Handle:      handle,
			SQLBuilder:  sqlBuilder,
			SQLExecutor: sqlExecutor,
			Clock:       clock,
		}
		return &Runnable{
			Context:            ctx,
			Store:              store,
			AppContextResolver: appContextResolver,
			UserServiceFactory: userServiceFactory,
			Logger:             runnableLogger,
		}
	}
	return factory
}
