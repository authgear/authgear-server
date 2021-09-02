package analytic

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(GlobalDBStore), "*"),
	wire.Struct(new(AppDBStore), "*"),
)
