package analytic

import (
	"github.com/authgear/authgear-server/pkg/lib/analytic"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/analyticredis"
	"github.com/authgear/authgear-server/pkg/lib/meter"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/log"

	"github.com/google/wire"
)

func NewLoggerFactory() *log.Factory {
	return log.NewFactory(log.LevelInfo)
}

func NewGlobalDatabaseCredentials(dbCredentials *config.DatabaseCredentials) *config.GlobalDatabaseCredentialsEnvironmentConfig {
	return &config.GlobalDatabaseCredentialsEnvironmentConfig{
		DatabaseURL:    dbCredentials.DatabaseURL,
		DatabaseSchema: dbCredentials.DatabaseSchema,
	}
}

var DependencySet = wire.NewSet(
	NewLoggerFactory,
	clock.DependencySet,
	config.NewDefaultDatabaseEnvironmentConfig,
	NewGlobalDatabaseCredentials,
	config.NewDefaultRedisEnvironmentConfig,
	globaldb.DependencySet,
	appdb.NewHandle,
	appdb.DependencySet,
	auditdb.NewReadHandle,
	auditdb.NewWriteHandle,
	auditdb.DependencySet,
	analyticredis.NewHandle,
	meter.DependencySet,

	analytic.DependencySet,
	wire.Bind(new(analytic.ReadCounterStore), new(*meter.ReadStoreRedis)),
	wire.Bind(new(analytic.MeterAuditDBReadStore), new(*meter.AuditDBReadStore)),
)
