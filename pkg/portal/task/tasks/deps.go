package tasks

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewSendMessagesLogger,
	wire.Struct(new(SendMessagesTask), "*"),
)
