package libstripe

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/util/redisutil"
)

type StripeCache struct {
	*redisutil.Cache
}

func NewStripeCache() *StripeCache {
	return &StripeCache{
		Cache: &redisutil.Cache{},
	}
}

var DependencySet = wire.NewSet(
	NewLogger,
	NewClientAPI,
	NewStripeCache,
	wire.Struct(new(Service), "*"),
	wire.Bind(new(Cache), new(*StripeCache)),
)
