package messaging

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(Limits), "*"),
	wire.Struct(new(Sender), "*"),
)
