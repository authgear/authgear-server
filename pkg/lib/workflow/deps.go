package workflow

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(Context), "*"),
	wire.Struct(new(StoreImpl), "*"),
	wire.Struct(new(Service), "*"),
	wire.Bind(new(Store), new(*StoreImpl)),
)
