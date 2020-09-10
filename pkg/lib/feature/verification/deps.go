package verification

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	wire.Struct(new(StoreRedis), "*"),
	wire.Bind(new(CodeStore), new(*StoreRedis)),
	wire.Struct(new(StorePQ), "*"),
	wire.Bind(new(ClaimStore), new(*StorePQ)),
	NewLogger,
	wire.Struct(new(Service), "*"),
	wire.Struct(new(CodeSender), "*"),
)
