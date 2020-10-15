package endpoint

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(RequestOriginProvider), "*"),
	wire.Bind(new(OriginProvider), new(*RequestOriginProvider)),
	wire.Struct(new(EndpointsProvider), "*"),
)
