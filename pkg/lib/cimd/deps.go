package cimd

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	ProvideCIMDHTTPClients,
	wire.Struct(new(Fetcher), "*"),

	wire.Struct(new(FetchSingleFlight), "*"),

	wire.Struct(new(RateLimiter), "*"),

	wire.Struct(new(Service), "*"),
)
