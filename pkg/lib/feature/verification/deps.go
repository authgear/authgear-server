package verification

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	wire.Struct(new(StorePQ), "*"),
	wire.Bind(new(ClaimStore), new(*StorePQ)),
	wire.Struct(new(Service), "*"),
)
