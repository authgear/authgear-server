package deps

import (
	"context"
	"net/http"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
)

func ProvideRequestContext(r *http.Request) context.Context { return r.Context() }

func ProvideConfigSource(ctrl *configsource.Controller) *configsource.ConfigSource {
	return ctrl.ForServer(configsource.ServerTypeMain)
}

type ConfigGetter struct {
	Context      context.Context
	Request      *http.Request
	ConfigSource *configsource.ConfigSource
}

func (g *ConfigGetter) GetConfig() (*config.Config, error) {
	return g.ConfigSource.ProvideConfig(g.Context, g.Request)
}

var DependencySet = wire.NewSet(
	wire.FieldsOf(new(*RootProvider),
		"EnvironmentConfig",
		"ConfigSourceConfig",
		"AuthgearConfig",
		"SentryHub",
		"LoggerFactory",
		"ConfigSourceController",
	),
	wire.FieldsOf(new(*config.EnvironmentConfig),
		"TrustProxy",
		"DevMode",
		"SentryDSN",
	),
	wire.FieldsOf(new(*RequestProvider),
		"RootProvider",
		"Request",
	),
	ProvideRequestContext,
	ProvideConfigSource,
	wire.Struct(new(ConfigGetter), "*"),
)
