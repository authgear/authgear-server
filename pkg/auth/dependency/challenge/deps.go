package challenge

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.NewSet(new(Provider), "*"),
)
