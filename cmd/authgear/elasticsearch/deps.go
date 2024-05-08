package elasticsearch

import (
	"github.com/google/wire"

	identityloginid "github.com/authgear/authgear-server/pkg/lib/authn/identity/loginid"
	identityoauth "github.com/authgear/authgear-server/pkg/lib/authn/identity/oauth"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/lib/rolesgroups"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/log"
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

func NewEmptyIdentityConfig() *config.IdentityConfig {
	return &config.IdentityConfig{
		OAuth: &config.OAuthSSOConfig{},
	}
}

var DependencySet = wire.NewSet(
	NewLoggerFactory,
	config.NewDefaultDatabaseEnvironmentConfig,
	NewGlobalDatabaseCredentials,
	NewEmptyIdentityConfig,
	NewReindexedTimestamps,
	globaldb.DependencySet,
	appdb.NewHandle,
	appdb.DependencySet,
	clock.DependencySet,
	wire.Struct(new(user.Store), "*"),
	wire.Struct(new(identityoauth.Store), "*"),
	wire.Struct(new(identityloginid.Store), "*"),
	wire.Struct(new(rolesgroups.Store), "*"),
	wire.Struct(new(configsource.Store), "*"),
	wire.Struct(new(AppLister), "*"),
	wire.Struct(new(Reindexer), "*"),
)
