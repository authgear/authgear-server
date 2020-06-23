package middlewares

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(CORSMiddleware), "*"),
	wire.Struct(new(RecoverMiddleware), "*"),
)
