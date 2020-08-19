package deps

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.FieldsOf(new(*RootProvider),
		"ServerConfig",
		"SentryHub",
		"LoggerFactory",
	),
)
