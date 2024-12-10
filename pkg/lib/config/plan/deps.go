package plan

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewStoreFactory,
	wire.Struct(new(Store), "*"),
)
