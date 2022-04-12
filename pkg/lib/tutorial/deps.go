package tutorial

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(Service), "*"),
	wire.Struct(new(StoreImpl), "*"),
	wire.Struct(new(Sink), "*"),
	wire.Bind(new(Store), new(*StoreImpl)),
)
