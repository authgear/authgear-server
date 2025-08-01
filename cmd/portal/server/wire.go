//go:build wireinject
// +build wireinject

package server

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

func newConfigSourceController(p *deps.RootProvider) *configsource.Controller {
	panic(wire.Build(
		clock.DependencySet,
		globaldb.DependencySet,
		configsource.NewResolveAppIDTypeDomain,
		configsource.DependencySet,
		configsource.ControllerDependencySet,
		wire.FieldsOf(new(*deps.RootProvider),
			"EnvironmentConfig",
			"ConfigSourceConfig",
			"AppBaseResources",
			"Database",
		),
		wire.FieldsOf(new(*config.EnvironmentConfig),
			"TrustProxy",
			"GlobalDatabase",
			"DatabaseConfig",
		),
	))
}
