package cimd

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	ProvideCIMDHTTPClients,
	wire.Struct(new(Fetcher), "*"),
)
