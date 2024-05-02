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
	NewWebHookLogger,
	NewDenoHookLogger,
	wire.Struct(new(Sink), "*"),
	wire.Bind(new(WebHook), new(*WebHookImpl)),
	wire.Struct(new(WebHookImpl), "*"),
	wire.Bind(new(EventWebHook), new(*EventWebHookImpl)),
	wire.Struct(new(EventWebHookImpl), "*"),
	wire.Struct(new(DenoHook), "*"),
	wire.Bind(new(EventDenoHook), new(*EventDenoHookImpl)),
	wire.Struct(new(EventDenoHookImpl), "*"),
)
