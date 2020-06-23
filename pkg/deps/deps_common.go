package deps

import (
	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	authredis "github.com/skygeario/skygear-server/pkg/auth/dependency/auth/redis"
	authenticatorbearertoken "github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/bearertoken"
	authenticatoroob "github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/oob"
	authenticatorpassword "github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/password"
	authenticatorprovider "github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/provider"
	authenticatorrecoverycode "github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/recoverycode"
	authenticatortotp "github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/totp"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/challenge"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	identityanonymous "github.com/skygeario/skygear-server/pkg/auth/dependency/identity/anonymous"
	identityloginid "github.com/skygeario/skygear-server/pkg/auth/dependency/identity/loginid"
	identityoauth "github.com/skygeario/skygear-server/pkg/auth/dependency/identity/oauth"
	identityprovider "github.com/skygeario/skygear-server/pkg/auth/dependency/identity/provider"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction"
	interactionflows "github.com/skygeario/skygear-server/pkg/auth/dependency/interaction/flows"
	interactionredis "github.com/skygeario/skygear-server/pkg/auth/dependency/interaction/redis"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth"
	oauthhandler "github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/handler"
	oauthpq "github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/pq"
	oauthredis "github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/redis"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oidc"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	sessionredis "github.com/skygeario/skygear-server/pkg/auth/dependency/session/redis"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/user"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/webapp"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/welcomemessage"
	"github.com/skygeario/skygear-server/pkg/clock"
	"github.com/skygeario/skygear-server/pkg/core/sentry"
	"github.com/skygeario/skygear-server/pkg/db"
	"github.com/skygeario/skygear-server/pkg/endpoints"
	"github.com/skygeario/skygear-server/pkg/task"
	taskqueue "github.com/skygeario/skygear-server/pkg/task/queue"
)

var commonDeps = wire.NewSet(
	configDeps,

	clock.DependencySet,
	db.DependencySet,
	sentry.DependencySet,

	wire.NewSet(
		challenge.DependencySet,
		wire.Bind(new(interactionflows.ChallengeProvider), new(*challenge.Provider)),
	),

	wire.NewSet(
		taskqueue.DependencySet,
		wire.Bind(new(task.Queue), new(*taskqueue.Queue)),
	),

	wire.NewSet(
		hook.DependencySet,
		wire.Bind(new(interaction.HookProvider), new(*hook.Provider)),
		wire.Bind(new(user.HookProvider), new(*hook.Provider)),
	),

	wire.NewSet(
		sessionredis.DependencySet,
		wire.Bind(new(session.Store), new(*sessionredis.Store)),

		session.DependencySet,
		wire.Bind(new(auth.IDPSessionResolver), new(*session.Resolver)),
		wire.Bind(new(oauth.ResolverSessionProvider), new(*session.Provider)),
		wire.Bind(new(oauthhandler.SessionProvider), new(*session.Provider)),
	),

	wire.NewSet(
		authredis.DependencySet,
		wire.Bind(new(auth.AccessEventStore), new(*authredis.EventStore)),

		auth.DependencySet,
		wire.Bind(new(session.AccessEventProvider), new(*auth.AccessEventProvider)),
	),

	wire.NewSet(
		authenticatorpassword.DependencySet,
		authenticatoroob.DependencySet,
		wire.Bind(new(interaction.OOBProvider), new(*authenticatoroob.Provider)),
		authenticatortotp.DependencySet,
		authenticatorbearertoken.DependencySet,
		authenticatorrecoverycode.DependencySet,

		authenticatorprovider.DependencySet,
		wire.Bind(new(authenticatorprovider.PasswordAuthenticatorProvider), new(*authenticatorpassword.Provider)),
		wire.Bind(new(authenticatorprovider.OOBOTPAuthenticatorProvider), new(*authenticatoroob.Provider)),
		wire.Bind(new(authenticatorprovider.TOTPAuthenticatorProvider), new(*authenticatortotp.Provider)),
		wire.Bind(new(authenticatorprovider.BearerTokenAuthenticatorProvider), new(*authenticatorbearertoken.Provider)),
		wire.Bind(new(authenticatorprovider.RecoveryCodeAuthenticatorProvider), new(*authenticatorrecoverycode.Provider)),

		wire.Bind(new(interaction.AuthenticatorProvider), new(*authenticatorprovider.Provider)),
	),

	wire.NewSet(
		identityloginid.DependencySet,
		identityoauth.DependencySet,
		identityanonymous.DependencySet,
		wire.Bind(new(interactionflows.AnonymousIdentityProvider), new(*identityanonymous.Provider)),

		identityprovider.DependencySet,
		wire.Bind(new(identityprovider.LoginIDIdentityProvider), new(*identityloginid.Provider)),
		wire.Bind(new(identityprovider.OAuthIdentityProvider), new(*identityoauth.Provider)),
		wire.Bind(new(identityprovider.AnonymousIdentityProvider), new(*identityanonymous.Provider)),
		wire.Bind(new(user.IdentityProvider), new(*identityprovider.Provider)),
		wire.Bind(new(interaction.IdentityProvider), new(*identityprovider.Provider)),
	),

	wire.NewSet(
		user.DependencySet,
		wire.Bind(new(auth.UserProvider), new(*user.Queries)),
		wire.Bind(new(interaction.UserProvider), new(*user.Provider)),
		wire.Bind(new(oidc.UserProvider), new(*user.Queries)),
		wire.Bind(new(hook.UserProvider), new(*user.RawProvider)),
	),

	wire.NewSet(
		welcomemessage.DependencySet,
		wire.Bind(new(user.WelcomeMessageProvider), new(*welcomemessage.Provider)),
	),

	wire.NewSet(
		oauthpq.DependencySet,
		wire.Bind(new(oauth.AuthorizationStore), new(*oauthpq.AuthorizationStore)),

		oauthredis.DependencySet,
		wire.Bind(new(oauth.AccessGrantStore), new(*oauthredis.GrantStore)),
		wire.Bind(new(oauth.CodeGrantStore), new(*oauthredis.GrantStore)),
		wire.Bind(new(oauth.OfflineGrantStore), new(*oauthredis.GrantStore)),

		oauth.DependencySet,
		wire.Bind(new(auth.AccessTokenSessionResolver), new(*oauth.Resolver)),
		wire.Bind(new(auth.AccessTokenSessionManager), new(*oauth.SessionManager)),
		wire.Bind(new(oauthhandler.OAuthURLProvider), new(*oauth.URLProvider)),
		wire.Value(oauthhandler.TokenGenerator(oauth.GenerateToken)),
	),

	wire.NewSet(
		oidc.DependencySet,
		wire.Value(oauthhandler.ScopesValidator(oidc.ValidateScopes)),
		wire.Bind(new(oauthhandler.IDTokenIssuer), new(*oidc.IDTokenIssuer)),
	),

	wire.NewSet(
		interactionredis.DependencySet,
		wire.Bind(new(interaction.Store), new(*interactionredis.Store)),

		interaction.DependencySet,
		wire.Bind(new(interactionflows.InteractionProvider), new(*interaction.Provider)),

		interactionflows.DependencySet,
		wire.Bind(new(webapp.AnonymousFlow), new(*interactionflows.AnonymousFlow)),
		wire.Bind(new(oauthhandler.AnonymousInteractionFlow), new(*interactionflows.AnonymousFlow)),
	),

	wire.NewSet(
		endpoints.DependencySet,
		wire.Bind(new(oauth.EndpointsProvider), new(*endpoints.Provider)),
		wire.Bind(new(webapp.EndpointsProvider), new(*endpoints.Provider)),
		wire.Bind(new(authenticatoroob.EndpointsProvider), new(*endpoints.Provider)),
		wire.Bind(new(oidc.EndpointsProvider), new(*endpoints.Provider)),
	),
)
