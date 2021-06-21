package middleware

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewCORSMiddlewareLogger,
	wire.Struct(new(CORSMiddleware), "*"),

	NewPanicLogMiddlewareLogger,
	wire.Struct(new(PanicLogMiddleware), "*"),

	wire.Struct(new(PanicWriteAPIResponseMiddleware), "*"),
	wire.Struct(new(PanicWriteEmptyResponseMiddleware), "*"),

	wire.Struct(new(PanicEndMiddleware), "*"),

	wire.Struct(new(SentryMiddleware), "*"),

	wire.Struct(new(BodyLimitMiddleware), "*"),
)
