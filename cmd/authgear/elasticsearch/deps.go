package elasticsearch

import (
	"github.com/google/wire"

	identityloginid "github.com/authgear/authgear-server/pkg/lib/authn/identity/loginid"
	identityoauth "github.com/authgear/authgear-server/pkg/lib/authn/identity/oauth"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	globaldb "github.com/authgear/authgear-server/pkg/lib/infra/db/global"
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
	appdb.NewHandle,
	appdb.DependencySet,
	wire.Struct(new(user.Store), "*"),
	wire.Struct(new(identityoauth.Store), "*"),
	wire.Struct(new(identityloginid.Store), "*"),
	wire.Struct(new(configsource.Store), "*"),
	wire.Struct(new(AppLister), "*"),
	wire.Struct(new(Reindexer), "*"),
)
