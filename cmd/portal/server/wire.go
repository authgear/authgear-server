//+build wireinject

package server

import (
	"context"
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

func newConfigSourceController(p *deps.RootProvider, c context.Context) *configsource.Controller {
	panic(wire.Build(
		clock.DependencySet,
		globaldb.DependencySet,
		configsource.DependencySet,
		wire.FieldsOf(new(*deps.RootProvider),
			"EnvironmentConfig",
			"ConfigSourceConfig",
			"LoggerFactory",
			"AppBaseResources",
			"Database",
		),
		wire.FieldsOf(new(*config.EnvironmentConfig),
			"TrustProxy",
			"Database",
		),
	))
}
