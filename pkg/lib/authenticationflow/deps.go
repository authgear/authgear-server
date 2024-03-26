package authenticationflow

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(Dependencies), "*"),
	NewServiceLogger,
	wire.Struct(new(Service), "*"),
	wire.Struct(new(StoreImpl), "*"),
	wire.Bind(new(Store), new(*StoreImpl)),
	wire.Struct(new(IntlMiddleware), "*"),
	wire.Struct(new(RateLimitMiddleware), "*"),
	NewWebsocketEventStore,
)
