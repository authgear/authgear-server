package workflow

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(Dependencies), "*"),
	NewServiceLogger,
	wire.Struct(new(SavePointImpl), "*"),
	wire.Bind(new(Savepoint), new(*SavePointImpl)),
	wire.Struct(new(Service), "*"),
	wire.Struct(new(StoreImpl), "*"),
	wire.Bind(new(Store), new(*StoreImpl)),
	wire.Bind(new(EventStore), new(*EventStoreImpl)),
	wire.Struct(new(ClientIDMiddleware), "*"),
	NewEventStore,
)
