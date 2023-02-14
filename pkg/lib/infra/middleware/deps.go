package middleware

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewCORSMiddlewareLogger,
	wire.Struct(new(CORSMiddleware), "*"),
	wire.Struct(new(CORSMatcher), "*"),

	NewPanicMiddlewareLogger,
	wire.Struct(new(PanicMiddleware), "*"),

	wire.Struct(new(SentryMiddleware), "*"),

	wire.Struct(new(BodyLimitMiddleware), "*"),
)
