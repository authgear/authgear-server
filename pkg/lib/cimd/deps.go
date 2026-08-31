package cimd

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/oauthclient"
)

var DependencySet = wire.NewSet(
	ProvideCIMDHTTPClients,
	wire.Struct(new(Fetcher), "*"),

	wire.Struct(new(FetchSingleFlight), "*"),

	ProvideNoopServiceRateLimiter,
	ProvideNoopServiceUsageLimiter,

	wire.Struct(new(Service), "*"),
	wire.Bind(new(ServiceOAuthClientCommands), new(*oauthclient.Commands)),
	wire.Bind(new(ServiceOAuthClientQueries), new(*oauthclient.Queries)),
	wire.Bind(new(ServiceDatabase), new(*appdb.Handle)),
)
