package elasticsearch

import (
	"github.com/google/wire"

	identityloginid "github.com/authgear/authgear-server/pkg/lib/authn/identity/loginid"
	identityoauth "github.com/authgear/authgear-server/pkg/lib/authn/identity/oauth"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/tenant"
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

var DependencySet = wire.NewSet(
	NewLoggerFactory,
	NewDatabaseConfig,
	tenant.NewHandle,
	tenant.NewPool,
	tenant.DependencySet,
	wire.Struct(new(user.Store), "*"),
	wire.Struct(new(identityoauth.Store), "*"),
	wire.Struct(new(identityloginid.Store), "*"),
	wire.Struct(new(Query), "*"),
)
