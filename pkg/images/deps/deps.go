package deps

import (
	"github.com/authgear/authgear-server/pkg/images/config"
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.FieldsOf(new(*RootProvider),
		"EnvironmentConfig",
		"LoggerFactory",
		"SentryHub",
	),
	wire.FieldsOf(new(*RequestProvider),
		"RootProvider",
	),
	wire.FieldsOf(new(*config.EnvironmentConfig),
		"TrustProxy",
	),
)
