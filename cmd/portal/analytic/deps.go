package analytic

import (
	"github.com/authgear/authgear-server/pkg/lib/analytic"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/analyticredis"
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

func NewRedisConfig() *config.RedisConfig {
	cfg := &config.RedisConfig{}
	cfg.SetDefaults()
	return cfg
}

var DependencySet = wire.NewSet(
	NewLoggerFactory,
	NewDatabaseConfig,
	NewDatabaseEnvironmentConfig,
	NewRedisConfig,
	globaldb.DependencySet,
	appdb.NewHandle,
	appdb.DependencySet,
	auditdb.NewReadHandle,
	auditdb.NewWriteHandle,
	auditdb.DependencySet,
	analyticredis.NewHandle,
	analytic.DependencySet,
)
