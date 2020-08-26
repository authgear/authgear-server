package service

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(AppService), "*"),
	wire.Struct(new(AdminAPIService), "*"),
)
