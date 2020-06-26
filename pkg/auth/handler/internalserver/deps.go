package internalserver

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewResolveHandlerLogger,
	wire.Struct(new(ResolveHandler), "*"),
)
