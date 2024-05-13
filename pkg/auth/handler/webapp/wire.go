//go:build wireinject
// +build wireinject

package webapp

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

func newGlobalSessionService(appID config.AppID, clock clock.Clock, redisHandle *appredis.Handle) *GlobalSessionService {
	panic(wire.Build(
		webapp.DependencySet,
		NewPublisher,
		wire.Struct(new(GlobalSessionService), "*"),
		wire.Bind(new(SessionStore), new(*webapp.SessionStoreRedis)),
	))
}
