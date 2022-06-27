package usage

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/analyticredis"
	"github.com/authgear/authgear-server/pkg/lib/meter"
	"github.com/authgear/authgear-server/pkg/lib/usage"
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
	config.NewDefaultDatabaseEnvironmentConfig,
	NewGlobalDatabaseCredentials,
	config.NewDefaultRedisEnvironmentConfig,
	globaldb.DependencySet,
	auditdb.DependencySet,
	auditdb.NewReadHandle,
	analyticredis.NewHandle,
	meter.DependencySet,

	usage.DependencySet,
	wire.Bind(new(usage.ReadCounterStore), new(*meter.ReadStoreRedis)),
	wire.Bind(new(usage.MeterAuditDBStore), new(*meter.AuditDBReadStore)),
)
