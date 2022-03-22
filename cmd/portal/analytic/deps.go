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

func NewGlobalDatabaseCredentials(dbCredentials *config.DatabaseCredentials) *config.GlobalDatabaseCredentialsEnvironmentConfig {
	return &config.GlobalDatabaseCredentialsEnvironmentConfig{
		DatabaseURL:    dbCredentials.DatabaseURL,
		DatabaseSchema: dbCredentials.DatabaseSchema,
	}
}

func NewRedisConfig() *config.RedisConfig {
	cfg := &config.RedisConfig{}
	cfg.SetDefaults()
	return cfg
}

var DependencySet = wire.NewSet(
	NewLoggerFactory,
	config.NewDefaultDatabaseEnvironmentConfig,
	NewGlobalDatabaseCredentials,
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
