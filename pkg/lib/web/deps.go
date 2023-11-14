package web

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(StaticAssetResolver), "*"),
	wire.Struct(new(ResponseRenderer), "*"),
	NewErrorCookieDef,
	wire.Struct(new(ErrorCookie), "*"),
)
