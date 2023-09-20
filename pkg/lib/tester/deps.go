package tester

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(TesterStore), "*"),
)
