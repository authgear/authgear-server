//go:build wireinject
// +build wireinject

package server

import (
	"github.com/google/wire"

	imagesconfig "github.com/authgear/authgear-server/pkg/images/config"
	"github.com/authgear/authgear-server/pkg/images/deps"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

var configSourceConfigDependencySet = wire.NewSet(
	globaldb.DependencySet,
	clock.DependencySet,
	wire.FieldsOf(new(*deps.RootProvider),
		"EnvironmentConfig",
		"LoggerFactory",
		"DatabasePool",
		"BaseResources",
	),
	wire.FieldsOf(new(*imagesconfig.EnvironmentConfig),
		"TrustProxy",
		"ConfigSource",
		"GlobalDatabase",
		"DatabaseConfig",
	),
)

func newConfigSourceController(p *deps.RootProvider) *configsource.Controller {
	panic(wire.Build(
		configSourceConfigDependencySet,
		configsource.NewResolveAppIDTypePath,
		configsource.DependencySet,
		configsource.ControllerDependencySet,
	))
}
