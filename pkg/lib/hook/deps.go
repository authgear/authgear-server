package hook

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewSyncHTTPClient,
	NewAsyncHTTPClient,
	NewSyncDenoClient,
	NewAsyncDenoClient,
	NewLogger,
	wire.Struct(new(Sink), "*"),
	wire.Bind(new(WebHook), new(*WebHookImpl)),
	wire.Struct(new(WebHookImpl), "*"),
	wire.Bind(new(DenoHook), new(*DenoHookImpl)),
	wire.Struct(new(DenoHookImpl), "*"),
)
