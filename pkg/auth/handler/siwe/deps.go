package siwe

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewNonceHandlerLogger,
	wire.Struct(new(NonceHandler), "*"),
)
