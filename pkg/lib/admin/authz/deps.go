package authz

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewLogger,
	wire.Struct(new(Middleware), "*"),
	wire.Struct(new(AuthzAdder), "*"),
)
