package otp

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	wire.Struct(new(Service), "*"),
	wire.Struct(new(MessageSender), "*"),
	wire.Struct(new(CodeStoreRedis), "*"),
	wire.Struct(new(MagicLinkStoreRedis), "*"),
	wire.Bind(new(CodeStore), new(*CodeStoreRedis)),
	wire.Bind(new(MagicLinkStore), new(*MagicLinkStoreRedis)),
	NewLogger,
)
