package otp

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	wire.Struct(new(Service), "*"),
	wire.Struct(new(MessageSender), "*"),
	wire.Struct(new(StoreRedis), "*"),
	wire.Bind(new(CodeStore), new(*StoreRedis)),
	NewLogger,
)
