package middleware

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(CORSMiddleware), "*"),

	NewRecoveryLogger,
	wire.Struct(new(RecoverMiddleware), "*"),

	NewLogPanicMiddlewareLogger,
	wire.Struct(new(LogPanicMiddleware), "*"),

	wire.Struct(new(PanicWriteAPIResponseMiddleware), "*"),

	wire.Struct(new(PanicEndMiddleware), "*"),

	wire.Struct(new(SentryMiddleware), "*"),

	wire.Struct(new(BodyLimitMiddleware), "*"),
)
