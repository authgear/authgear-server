package libstripe

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewLogger,
	NewClientAPI,
	wire.Struct(new(Service), "*"),
)
