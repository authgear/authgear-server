package transport

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(GraphQLHandler), "*"),
	wire.Struct(new(RuntimeConfigHandler), "*"),
	wire.Struct(new(AdminAPIHandler), "*"),
)
