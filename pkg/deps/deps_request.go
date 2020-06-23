package deps

import (
	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/webapp"
	"github.com/skygeario/skygear-server/pkg/middlewares"
)

var requestDeps = wire.NewSet(
	commonDeps,

	middlewares.DependencySet,
	webapp.DependencySet,
)
