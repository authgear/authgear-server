package deps

import (
	"context"
	"net/http"

	"github.com/google/wire"
)

func ProvideRequestContext(r *http.Request) context.Context { return r.Context() }

var DependencySet = wire.NewSet(
	wire.FieldsOf(new(*RootProvider),
		"ServerConfig",
		"SentryHub",
		"LoggerFactory",
	),
	wire.FieldsOf(new(*RequestProvider),
		"RootProvider",
		"Request",
	),
	ProvideRequestContext,
)
