package oob

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(Store), "*"),
	wire.Struct(new(Provider), "*"),
	wire.Struct(new(CodeSender), "*"),
	wire.Struct(new(StoreRedis), "*"),
	wire.Bind(new(CodeStore), new(*StoreRedis)),
	NewLogger,
)
