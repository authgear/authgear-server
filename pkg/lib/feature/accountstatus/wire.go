//go:build wireinject
// +build wireinject

package accountstatus

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/util/backgroundjob"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

func newRunnable(
	pool *db.Pool,
	globalDBCredentials *config.GlobalDatabaseCredentialsEnvironmentConfig,
	databaseCfg *config.DatabaseEnvironmentConfig,
	clock clock.Clock,
	appContextResolver AppContextResolver,
	userServiceFactory UserServiceFactory,
) backgroundjob.Runnable {
	panic(wire.Build(RunnableDependencySet))
}
