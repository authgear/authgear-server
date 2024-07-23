package dpop

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(Provider), "*"),
	wire.Struct(new(Middleware), "*"),
)
