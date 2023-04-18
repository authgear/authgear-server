package usage

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewLogger,
	wire.Struct(new(GlobalDBStore), "*"),
	wire.Struct(new(CountCollector), "*"),
	wire.Struct(new(Limiter), "*"),
)
