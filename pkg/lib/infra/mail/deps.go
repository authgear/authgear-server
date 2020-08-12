package mail

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewGomailDialer,
	NewLogger,
	wire.Struct(new(Sender), "*"),
)
