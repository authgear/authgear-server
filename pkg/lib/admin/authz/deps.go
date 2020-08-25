package authz

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewAuthzLogger,
	wire.Struct(new(Middleware), "*"),
)
