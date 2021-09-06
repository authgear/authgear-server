package analytic

import (
	"github.com/authgear/authgear-server/pkg/lib/analytic"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/google/wire"
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
	appdb.NewHandle,
	appdb.DependencySet,
	analytic.DependencySet,
)
