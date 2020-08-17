package loginid

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(TypeCheckerFactory), "*"),
	wire.Struct(new(Checker), "*"),
	wire.Struct(new(NormalizerFactory), "*"),
	wire.Struct(new(Store), "*"),
	wire.Struct(new(Provider), "*"),
)
