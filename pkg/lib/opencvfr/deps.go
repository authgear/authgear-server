package opencvfr

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	wire.Struct(new(Store), "*"),
	wire.Struct(new(Service), "*"),
)
