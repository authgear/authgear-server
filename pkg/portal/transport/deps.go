package transport

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(GraphQLHandler), "*"),
	wire.Struct(new(SystemConfigHandler), "*"),
	wire.Struct(new(AdminAPIHandler), "*"),
	wire.Struct(new(StaticAssetsHandler), "*"),
	wire.Struct(new(StripeWebhookHandler), "*"),
	wire.Struct(new(OsanoHandler), "*"),
)
