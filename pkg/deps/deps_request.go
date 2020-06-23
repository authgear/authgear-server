package deps

import (
	"github.com/google/wire"

	identityanonymous "github.com/skygeario/skygear-server/pkg/auth/dependency/identity/anonymous"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/webapp"
	handlersession "github.com/skygeario/skygear-server/pkg/auth/handler/session"
	"github.com/skygeario/skygear-server/pkg/middlewares"
)

var requestDeps = wire.NewSet(
	wire.NewSet(
		commonDeps,

		middlewares.DependencySet,
		webapp.DependencySet,
	),

	handlersession.DependencySet,
	wire.Bind(new(handlersession.AnonymousIdentityProvider), new(*identityanonymous.Provider)),
)
