package messaging

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewLogger,
	wire.Struct(new(Limits), "*"),
	wire.Struct(new(Sender), "*"),
)
