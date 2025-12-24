package service

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(Store), "*"),
	wire.Struct(new(RateLimits), "*"),
	wire.Struct(new(Lockout), "*"),
	wire.Struct(new(ReadOnlyService), "*"),
	wire.Struct(new(Service), "*"),
)
