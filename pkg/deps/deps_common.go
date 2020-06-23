package deps

import (
	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	authredis "github.com/skygeario/skygear-server/pkg/auth/dependency/auth/redis"
	authenticatorpassword "github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/password"
	identityanonymous "github.com/skygeario/skygear-server/pkg/auth/dependency/identity/anonymous"
	identityloginid "github.com/skygeario/skygear-server/pkg/auth/dependency/identity/loginid"
	identityoauth "github.com/skygeario/skygear-server/pkg/auth/dependency/identity/oauth"
	identityprovider "github.com/skygeario/skygear-server/pkg/auth/dependency/identity/provider"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	sessionredis "github.com/skygeario/skygear-server/pkg/auth/dependency/session/redis"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/user"
	"github.com/skygeario/skygear-server/pkg/clock"
	"github.com/skygeario/skygear-server/pkg/core/sentry"
	"github.com/skygeario/skygear-server/pkg/db"
)

var commonDeps = wire.NewSet(
	configDeps,

	clock.DependencySet,
	db.DependencySet,
	sentry.DependencySet,

	wire.NewSet(
		sessionredis.DependencySet,
		wire.Bind(new(session.Store), new(*sessionredis.Store)),

		session.DependencySet,
		wire.Bind(new(auth.IDPSessionResolver), new(*session.Resolver)),
		wire.Bind(new(auth.AccessTokenSessionResolver), new(*session.Resolver)),
	),

	wire.NewSet(
		authredis.DependencySet,
		wire.Bind(new(auth.AccessEventStore), new(*authredis.EventStore)),

		auth.DependencySet,
		wire.Bind(new(session.AccessEventProvider), new(*auth.AccessEventProvider)),
	),

	wire.NewSet(
		authenticatorpassword.DependencySet,
	),

	wire.NewSet(
		identityloginid.DependencySet,
		identityoauth.DependencySet,
		identityanonymous.DependencySet,
		identityprovider.DependencySet,
		wire.Bind(new(identityprovider.LoginIDIdentityProvider), new(*identityloginid.Provider)),
		wire.Bind(new(identityprovider.OAuthIdentityProvider), new(*identityoauth.Provider)),
		wire.Bind(new(identityprovider.AnonymousIdentityProvider), new(*identityanonymous.Provider)),

		wire.Bind(new(user.IdentityProvider), new(*identityprovider.Provider)),
	),

	wire.NewSet(
		user.DependencySet,
		wire.Bind(new(auth.UserProvider), new(*user.Queries)),
	),
)
