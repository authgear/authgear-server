package session

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewMiddlewareLogger,
	wire.Struct(new(Middleware), "*"),
	wire.Struct(new(Manager), "*"),
	NewSessionCookieDef,
)
