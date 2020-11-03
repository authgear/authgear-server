package webapp

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(URLProvider), "*"),
	wire.Struct(new(AuthenticateURLProvider), "*"),

	NewCSRFCookieDef,
	NewUATokenCookieDef,
	NewErrorCookieDef,
	wire.Struct(new(ErrorCookie), "*"),

	wire.Struct(new(CSPMiddleware), "*"),
	wire.Struct(new(CSRFMiddleware), "*"),
	wire.Struct(new(AuthEntryPointMiddleware), "*"),
	wire.Struct(new(StateMiddleware), "*"),
	wire.Bind(new(StateMiddlewareStates), new(*RedisStore)),

	wire.Struct(new(RedisStore), "*"),
	wire.Bind(new(Store), new(*RedisStore)),
	wire.Struct(new(Service), "*"),
	wire.Bind(new(PageService), new(*Service)),
	NewServiceLogger,
)
