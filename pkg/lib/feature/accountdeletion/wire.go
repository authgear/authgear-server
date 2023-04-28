//go:build wireinject
// +build wireinject

package accountdeletion

import (
	"context"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/util/backgroundjob"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/log"
)

func newRunnable(
	context context.Context,
	pool *db.Pool,
	globalDBCredentials *config.GlobalDatabaseCredentialsEnvironmentConfig,
	databaseCfg *config.DatabaseEnvironmentConfig,
	logFactory *log.Factory,
	clock clock.Clock,
	appContextResolver AppContextResolver,
	userServiceFactory UserServiceFactory,
) backgroundjob.Runnable {
	panic(wire.Build(RunnableDependencySet))
}
