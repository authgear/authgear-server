//+build wireinject

package server

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

func newConfigSourceController(p *deps.RootProvider) *configsource.Controller {
	panic(wire.Build(
		clock.DependencySet,
		configsource.DependencySet,
		wire.FieldsOf(new(*deps.RootProvider),
			"EnvironmentConfig",
			"ConfigSourceConfig",
			"LoggerFactory",
			"AppBaseResources",
		),
		wire.FieldsOf(new(*config.EnvironmentConfig),
			"TrustProxy",
		),
	))
}
