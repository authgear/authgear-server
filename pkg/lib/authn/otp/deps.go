package otp

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	wire.Struct(new(Service), "*"),
	wire.Struct(new(MessageSender), "*"),
	wire.Struct(new(CodeStoreRedis), "*"),
	wire.Struct(new(LoginLinkStoreRedis), "*"),
	wire.Bind(new(CodeStore), new(*CodeStoreRedis)),
	wire.Bind(new(LoginLinkStore), new(*LoginLinkStoreRedis)),
	NewLogger,
)
