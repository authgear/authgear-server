package workflow

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(Dependencies), "*"),
	wire.Struct(new(Service), "*"),
	wire.Struct(new(StoreImpl), "*"),
	wire.Bind(new(Store), new(*StoreImpl)),
	wire.Bind(new(EventStore), new(*EventStoreImpl)),
	wire.Struct(new(IntlMiddleware), "*"),
	NewEventStore,
)
