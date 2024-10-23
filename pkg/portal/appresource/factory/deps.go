package factory

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewManagerFactoryLogger,
	wire.Struct(new(ManagerFactory), "*"),
)
