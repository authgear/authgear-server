package transport

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(GraphQLHandler), "*"),
	NewAuthorizationMiddlewareLogger,
	wire.Struct(new(AuthorizationMiddleware), "*"),
)
