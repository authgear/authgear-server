package middleware

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(CORSMiddleware), "*"),
	NewRecoveryLogger,
	wire.Struct(new(RecoverMiddleware), "*"),
	wire.Struct(new(SentryMiddleware), "*"),
)
