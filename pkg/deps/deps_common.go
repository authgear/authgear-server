package deps

import (
	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/core/sentry"
)

var commonDeps = wire.NewSet(
	configDeps,

	sentry.DependencySet,
)
