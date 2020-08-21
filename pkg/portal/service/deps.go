package service

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewLibConfig,
	wire.Struct(new(AppService), "*"),
)
