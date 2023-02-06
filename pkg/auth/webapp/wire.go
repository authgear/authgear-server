//go:build wireinject
// +build wireinject

package webapp

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/google/wire"
)

func newGlobalSessionService(appID config.AppID, clock clock.Clock, redisHandle *appredis.Handle) *GlobalSessionService {
	panic(wire.Build(
		DependencySet,
		wire.Struct(new(GlobalSessionService), "*"),
		wire.Bind(new(WebSessionStore), new(*SessionStoreRedis)),
	))
}
