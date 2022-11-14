package hook

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewSyncHTTPClient,
	NewAsyncHTTPClient,
	NewLogger,
	wire.Struct(new(Sink), "*"),
)
