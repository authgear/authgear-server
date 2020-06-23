package deps

import (
	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth/redis"
	authenticatorpassword "github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/password"
	"github.com/skygeario/skygear-server/pkg/clock"
	"github.com/skygeario/skygear-server/pkg/core/sentry"
	"github.com/skygeario/skygear-server/pkg/db"
)

var commonDeps = wire.NewSet(
	configDeps,

	clock.DependencySet,
	db.DependencySet,
	redis.DependencySet,
	sentry.DependencySet,
	authenticatorpassword.DependencySet,
)
