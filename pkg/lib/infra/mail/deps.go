package mail

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewGomailDialer,
	wire.Struct(new(Sender), "*"),
)
