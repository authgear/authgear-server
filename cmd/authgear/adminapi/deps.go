package adminapi

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/admin/authz"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/log"
)

func NewLoggerFactory() *log.Factory {
	return log.NewFactory(log.LevelInfo)
}

var DependencySet = wire.NewSet(
	wire.Struct(new(Invoker), "*"),

	NewLoggerFactory,
	clock.DependencySet,
	globaldb.DependencySet,
	config.NewDefaultDatabaseEnvironmentConfig,
	authz.DependencySet,
	wire.Struct(new(configsource.Store), "*"),
)
