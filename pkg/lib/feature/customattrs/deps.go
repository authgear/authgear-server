package customattrs

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(ServiceNoEvent), "*"),
)
