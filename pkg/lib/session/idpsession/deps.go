package idpsession

import (
	"github.com/google/wire"

	corerand "github.com/authgear/authgear-server/pkg/util/rand"
)

var DependencySet = wire.NewSet(
	NewStoreRedisLogger,
	wire.Struct(new(StoreRedis), "*"),
	wire.Bind(new(Store), new(*StoreRedis)),

	wire.Value(Rand(corerand.SecureRand)),
	wire.Struct(new(Provider), "*"),
	wire.Struct(new(Resolver), "*"),
	wire.Struct(new(Manager), "*"),
	wire.Bind(new(resolverProvider), new(*Provider)),
)
