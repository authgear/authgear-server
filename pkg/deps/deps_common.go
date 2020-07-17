package deps

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/auth/dependency/auth"
	authredis "github.com/authgear/authgear-server/pkg/auth/dependency/auth/redis"
	authenticatorbearertoken "github.com/authgear/authgear-server/pkg/auth/dependency/authenticator/bearertoken"
	authenticatoroob "github.com/authgear/authgear-server/pkg/auth/dependency/authenticator/oob"
	authenticatorpassword "github.com/authgear/authgear-server/pkg/auth/dependency/authenticator/password"
	authenticatorprovider "github.com/authgear/authgear-server/pkg/auth/dependency/authenticator/provider"
	authenticatorrecoverycode "github.com/authgear/authgear-server/pkg/auth/dependency/authenticator/recoverycode"
	authenticatortotp "github.com/authgear/authgear-server/pkg/auth/dependency/authenticator/totp"
	"github.com/authgear/authgear-server/pkg/auth/dependency/challenge"
	"github.com/authgear/authgear-server/pkg/auth/dependency/forgotpassword"
	"github.com/authgear/authgear-server/pkg/auth/dependency/hook"
	identityanonymous "github.com/authgear/authgear-server/pkg/auth/dependency/identity/anonymous"
	identityloginid "github.com/authgear/authgear-server/pkg/auth/dependency/identity/loginid"
	identityoauth "github.com/authgear/authgear-server/pkg/auth/dependency/identity/oauth"
	identityprovider "github.com/authgear/authgear-server/pkg/auth/dependency/identity/provider"
	"github.com/authgear/authgear-server/pkg/auth/dependency/interaction"
	interactionflows "github.com/authgear/authgear-server/pkg/auth/dependency/interaction/flows"
	"github.com/authgear/authgear-server/pkg/auth/dependency/oauth"
	oauthhandler "github.com/authgear/authgear-server/pkg/auth/dependency/oauth/handler"
	oauthpq "github.com/authgear/authgear-server/pkg/auth/dependency/oauth/pq"
	oauthredis "github.com/authgear/authgear-server/pkg/auth/dependency/oauth/redis"
	"github.com/authgear/authgear-server/pkg/auth/dependency/oidc"
	"github.com/authgear/authgear-server/pkg/auth/dependency/session"
	sessionredis "github.com/authgear/authgear-server/pkg/auth/dependency/session/redis"
	"github.com/authgear/authgear-server/pkg/auth/dependency/sso"
	"github.com/authgear/authgear-server/pkg/auth/dependency/user"
	"github.com/authgear/authgear-server/pkg/auth/dependency/verification"
	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
	"github.com/authgear/authgear-server/pkg/auth/dependency/welcomemessage"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/clock"
	"github.com/authgear/authgear-server/pkg/core/sentry"
	"github.com/authgear/authgear-server/pkg/db"
	"github.com/authgear/authgear-server/pkg/endpoints"
	"github.com/authgear/authgear-server/pkg/otp"
	"github.com/authgear/authgear-server/pkg/task"
	taskqueue "github.com/authgear/authgear-server/pkg/task/queue"
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
		wire.Bind(new(auth.HookProvider), new(*hook.Provider)),
		wire.Bind(new(forgotpassword.HookProvider), new(*hook.Provider)),
		wire.Bind(new(interactionflows.HookProvider), new(*hook.Provider)),
	),

	wire.NewSet(
		sessionredis.DependencySet,
		wire.Bind(new(session.Store), new(*sessionredis.Store)),

		session.DependencySet,
		wire.Bind(new(auth.IDPSessionResolver), new(*session.Resolver)),
		wire.Bind(new(auth.IDPSessionManager), new(*session.Manager)),
		wire.Bind(new(oauth.ResolverSessionProvider), new(*session.Provider)),
		wire.Bind(new(oauthhandler.SessionProvider), new(*session.Provider)),
		wire.Bind(new(interactionflows.SessionProvider), new(*session.Provider)),
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
		wire.Bind(new(verification.AuthenticatorProvider), new(*authenticatorprovider.Provider)),
	),

	wire.NewSet(
		identityloginid.DependencySet,
		wire.Bind(new(sso.LoginIDNormalizerFactory), new(*identityloginid.NormalizerFactory)),
		wire.Bind(new(forgotpassword.LoginIDProvider), new(*identityloginid.Provider)),
		identityoauth.DependencySet,
		identityanonymous.DependencySet,
		wire.Bind(new(interactionflows.AnonymousIdentityProvider), new(*identityanonymous.Provider)),

		identityprovider.DependencySet,
		wire.Bind(new(identityprovider.LoginIDIdentityProvider), new(*identityloginid.Provider)),
		wire.Bind(new(identityprovider.OAuthIdentityProvider), new(*identityoauth.Provider)),
		wire.Bind(new(identityprovider.AnonymousIdentityProvider), new(*identityanonymous.Provider)),
		wire.Bind(new(user.IdentityProvider), new(*identityprovider.Provider)),
		wire.Bind(new(interaction.IdentityProvider), new(*identityprovider.Provider)),
		wire.Bind(new(interactionflows.IdentityProvider), new(*identityprovider.Provider)),
		wire.Bind(new(verification.IdentityProvider), new(*identityprovider.Provider)),
	),

	wire.NewSet(
		user.DependencySet,
		wire.Bind(new(auth.UserProvider), new(*user.Queries)),
		wire.Bind(new(interaction.UserProvider), new(*user.Provider)),
		wire.Bind(new(interactionflows.UserProvider), new(*user.Provider)),
		wire.Bind(new(forgotpassword.UserProvider), new(*user.Queries)),
		wire.Bind(new(oidc.UserProvider), new(*user.Queries)),
		wire.Bind(new(hook.UserProvider), new(*user.RawProvider)),
	),

	wire.NewSet(
		welcomemessage.DependencySet,
		wire.Bind(new(user.WelcomeMessageProvider), new(*welcomemessage.Provider)),
	),

	wire.NewSet(
		forgotpassword.DependencySet,
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
		interaction.DependencySet,
		wire.Bind(new(interactionflows.InteractionProvider), new(*interaction.Provider)),

		interactionflows.DependencySet,
		wire.Bind(new(webapp.AnonymousFlow), new(*interactionflows.AnonymousFlow)),
		wire.Bind(new(webapp.ResponderInteractions), new(*interaction.Provider)),
		wire.Bind(new(oauthhandler.AnonymousInteractionFlow), new(*interactionflows.AnonymousFlow)),
		wire.Bind(new(forgotpassword.ResetPasswordFlow), new(*interactionflows.PasswordFlow)),
	),

	wire.NewSet(
		endpoints.DependencySet,
		wire.Bind(new(oauth.EndpointsProvider), new(*endpoints.Provider)),
		wire.Bind(new(webapp.EndpointsProvider), new(*endpoints.Provider)),
		wire.Bind(new(handlerwebapp.KeyURIImageEndpoints), new(*endpoints.Provider)),
		wire.Bind(new(authenticatoroob.EndpointsProvider), new(*endpoints.Provider)),
		wire.Bind(new(oidc.EndpointsProvider), new(*endpoints.Provider)),
		wire.Bind(new(sso.EndpointsProvider), new(*endpoints.Provider)),
		wire.Bind(new(otp.EndpointsProvider), new(*endpoints.Provider)),
	),

	wire.NewSet(
		verification.DependencySet,
		wire.Bind(new(user.VerificationService), new(*verification.Service)),
	),

	wire.NewSet(
		otp.DependencySet,
		wire.Bind(new(authenticatoroob.OTPMessageSender), new(*otp.MessageSender)),
		wire.Bind(new(verification.OTPMessageSender), new(*otp.MessageSender)),
	),
)
