package time

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	NewProvider,
)
