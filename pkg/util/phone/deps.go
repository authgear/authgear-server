package phone

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	wire.Struct(new(LegalParser), "*"),
	wire.Struct(new(LegalAndValidParser), "*"),
)
