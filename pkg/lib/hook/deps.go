package hook

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewSyncHTTPClient,
	NewAsyncHTTPClient,
	NewLogger,
	wire.Struct(new(Sink), "*"),
	wire.Bind(new(WebHook), new(*WebHookImpl)),
	wire.Struct(new(WebHookImpl), "*"),
)
