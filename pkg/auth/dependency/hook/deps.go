package hook

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(Provider), "*"),
	wire.Struct(new(Deliverer), "*"),
	wire.Struct(new(Mutator), "*"),
	wire.Struct(new(Store), "*"),
)
