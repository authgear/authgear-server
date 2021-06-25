package plan

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/portal/lib/plan"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/log"
)

func NewLoggerFactory() *log.Factory {
	return log.NewFactory(log.LevelInfo)
}

func NewDatabaseConfig() *config.DatabaseConfig {
	cfg := &config.DatabaseConfig{}
	cfg.SetDefaults()
	return cfg
}

func NewDatabaseEnvironmentConfig(dbCredentials *config.DatabaseCredentials, dbConfig *config.DatabaseConfig) *config.DatabaseEnvironmentConfig {
	return &config.DatabaseEnvironmentConfig{
		DatabaseURL:            dbCredentials.DatabaseURL,
		DatabaseSchema:         dbCredentials.DatabaseSchema,
		MaxOpenConn:            *dbConfig.MaxOpenConnection,
		MaxIdleConn:            *dbConfig.MaxIdleConnection,
		ConnMaxLifetimeSeconds: int(*dbConfig.MaxConnectionLifetime),
		ConnMaxIdleTimeSeconds: int(*dbConfig.IdleConnectionTimeout),
	}
}

var DependencySet = wire.NewSet(
	NewLoggerFactory,
	NewDatabaseConfig,
	NewDatabaseEnvironmentConfig,
	globaldb.DependencySet,
	clock.DependencySet,
	plan.DependencySet,
	wire.Struct(new(Service), "*"),
)
