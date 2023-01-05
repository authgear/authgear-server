package oob

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(Store), "*"),
	wire.Struct(new(Provider), "*"),
	wire.Struct(new(CodeSender), "*"),
)
