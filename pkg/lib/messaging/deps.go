package messaging

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewLogger,
	wire.Struct(new(RateLimits), "*"),
	wire.Struct(new(Sender), "*"),
)
