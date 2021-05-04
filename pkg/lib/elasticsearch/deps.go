package elasticsearch

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewClient,
	wire.Struct(new(Service), "*"),
)
