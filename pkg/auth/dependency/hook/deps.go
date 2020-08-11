package hook

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewSyncHTTPClient,
	NewAsyncHTTPClient,
	NewLogger,
	wire.Struct(new(Deliverer), "*"),
	wire.Bind(new(deliverer), new(*Deliverer)),
	wire.Struct(new(Store), "*"),
	wire.Bind(new(store), new(*Store)),
	wire.Struct(new(Provider), "*"),
)
