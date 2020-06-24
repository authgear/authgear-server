package task

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewPwHousekeeperLogger,
	wire.Struct(new(PwHousekeeperTask), "*"),
	NewSendMessagesLogger,
	wire.Struct(new(SendMessagesTask), "*"),
)
