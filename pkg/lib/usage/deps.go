package usage

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(HardSMSBucketer), "*"),
	wire.Struct(new(GlobalDBStore), "*"),
	wire.Struct(new(CountCollector), "*"),
)
