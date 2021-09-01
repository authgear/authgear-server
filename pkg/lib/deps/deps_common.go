package deps

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/audit"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	authenticatoroob "github.com/authgear/authgear-server/pkg/lib/authn/authenticator/oob"
	authenticatorpassword "github.com/authgear/authgear-server/pkg/lib/authn/authenticator/password"
	authenticatorservice "github.com/authgear/authgear-server/pkg/lib/authn/authenticator/service"
	authenticatortotp "github.com/authgear/authgear-server/pkg/lib/authn/authenticator/totp"
	"github.com/authgear/authgear-server/pkg/lib/authn/challenge"
	identityanonymous "github.com/authgear/authgear-server/pkg/lib/authn/identity/anonymous"
	identitybiometric "github.com/authgear/authgear-server/pkg/lib/authn/identity/biometric"
	identityloginid "github.com/authgear/authgear-server/pkg/lib/authn/identity/loginid"
	identityoauth "github.com/authgear/authgear-server/pkg/lib/authn/identity/oauth"
	identityservice "github.com/authgear/authgear-server/pkg/lib/authn/identity/service"
	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/authn/sso"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	libes "github.com/authgear/authgear-server/pkg/lib/elasticsearch"
	"github.com/authgear/authgear-server/pkg/lib/event"
	"github.com/authgear/authgear-server/pkg/lib/facade"
	"github.com/authgear/authgear-server/pkg/lib/feature/forgotpassword"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/feature/welcomemessage"
	"github.com/authgear/authgear-server/pkg/lib/healthz"
	"github.com/authgear/authgear-server/pkg/lib/hook"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	oauthhandler "github.com/authgear/authgear-server/pkg/lib/oauth/handler"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	oidchandler "github.com/authgear/authgear-server/pkg/lib/oauth/oidc/handler"
	oauthpq "github.com/authgear/authgear-server/pkg/lib/oauth/pq"
	oauthredis "github.com/authgear/authgear-server/pkg/lib/oauth/redis"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/access"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/lib/translation"
	"github.com/authgear/authgear-server/pkg/lib/web"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var CommonDependencySet = wire.NewSet(
	configDeps,
	utilsDeps,

	appdb.DependencySet,
	auditdb.DependencySet,
	template.DependencySet,

	healthz.DependencySet,

	wire.NewSet(
		authenticationinfo.DependencySet,
		wire.Bind(new(interaction.AuthenticationInfoService), new(*authenticationinfo.StoreRedis)),
		wire.Bind(new(oauthhandler.AuthenticationInfoService), new(*authenticationinfo.StoreRedis)),
	),

	wire.NewSet(
		libes.DependencySet,
		wire.Bind(new(interaction.SearchService), new(*libes.Service)),
	),

	wire.NewSet(
		challenge.DependencySet,
		wire.Bind(new(interaction.ChallengeProvider), new(*challenge.Provider)),
	),

	wire.NewSet(
		event.DependencySet,
		wire.Bind(new(interaction.EventService), new(*event.Service)),
		wire.Bind(new(user.EventService), new(*event.Service)),
		wire.Bind(new(session.EventService), new(*event.Service)),
		wire.Bind(new(otp.EventService), new(*event.Service)),
		wire.Bind(new(forgotpassword.EventService), new(*event.Service)),
	),

	wire.NewSet(
		hook.DependencySet,
	),

	wire.NewSet(
		audit.DependencySet,
	),

	wire.NewSet(
		idpsession.DependencySet,

		wire.Bind(new(session.IDPSessionResolver), new(*idpsession.Resolver)),
		wire.Bind(new(session.IDPSessionManager), new(*idpsession.Manager)),
		wire.Bind(new(oauth.ResolverSessionProvider), new(*idpsession.Provider)),
		wire.Bind(new(oauthhandler.SessionProvider), new(*idpsession.Provider)),
		wire.Bind(new(interaction.SessionProvider), new(*idpsession.Provider)),
		wire.Bind(new(interaction.SessionManager), new(*idpsession.Manager)),
		wire.Bind(new(facade.IDPSessionManager), new(*idpsession.Manager)),
	),

	wire.NewSet(
		access.DependencySet,
		session.DependencySet,
		wire.Bind(new(idpsession.AccessEventProvider), new(*access.EventProvider)),
		wire.Bind(new(oidchandler.LogoutSessionManager), new(*session.Manager)),
		wire.Bind(new(oauthhandler.SessionManager), new(*session.Manager)),
	),

	wire.NewSet(
		authenticatorpassword.DependencySet,
		wire.Bind(new(facade.PasswordHistoryStore), new(*authenticatorpassword.HistoryStore)),
		authenticatoroob.DependencySet,
		wire.Bind(new(interaction.OOBAuthenticatorProvider), new(*authenticatoroob.Provider)),
		wire.Bind(new(interaction.OOBCodeSender), new(*authenticatoroob.CodeSender)),
		authenticatortotp.DependencySet,

		authenticatorservice.DependencySet,
		wire.Bind(new(authenticatorservice.PasswordAuthenticatorProvider), new(*authenticatorpassword.Provider)),
		wire.Bind(new(authenticatorservice.OOBOTPAuthenticatorProvider), new(*authenticatoroob.Provider)),
		wire.Bind(new(authenticatorservice.TOTPAuthenticatorProvider), new(*authenticatortotp.Provider)),

		wire.Bind(new(facade.AuthenticatorService), new(*authenticatorservice.Service)),
	),

	wire.NewSet(
		mfa.DependencySet,

		wire.Bind(new(interaction.MFAService), new(*mfa.Service)),
		wire.Bind(new(facade.MFAService), new(*mfa.Service)),
	),

	wire.NewSet(
		identityloginid.DependencySet,
		wire.Bind(new(sso.LoginIDNormalizerFactory), new(*identityloginid.NormalizerFactory)),
		wire.Bind(new(interaction.LoginIDNormalizerFactory), new(*identityloginid.NormalizerFactory)),
		identityoauth.DependencySet,
		identityanonymous.DependencySet,
		wire.Bind(new(interaction.AnonymousIdentityProvider), new(*identityanonymous.Provider)),

		identitybiometric.DependencySet,
		wire.Bind(new(interaction.BiometricIdentityProvider), new(*identitybiometric.Provider)),

		identityservice.DependencySet,
		wire.Bind(new(identityservice.LoginIDIdentityProvider), new(*identityloginid.Provider)),
		wire.Bind(new(identityservice.OAuthIdentityProvider), new(*identityoauth.Provider)),
		wire.Bind(new(identityservice.AnonymousIdentityProvider), new(*identityanonymous.Provider)),
		wire.Bind(new(identityservice.BiometricIdentityProvider), new(*identitybiometric.Provider)),

		wire.Bind(new(facade.IdentityService), new(*identityservice.Service)),
	),

	wire.NewSet(
		facade.DependencySet,

		wire.Bind(new(interaction.IdentityService), new(facade.IdentityFacade)),
		wire.Bind(new(interaction.AuthenticatorService), new(facade.AuthenticatorFacade)),
		wire.Bind(new(forgotpassword.AuthenticatorService), new(facade.AuthenticatorFacade)),
		wire.Bind(new(user.IdentityService), new(facade.IdentityFacade)),
		wire.Bind(new(user.AuthenticatorService), new(facade.AuthenticatorFacade)),
		wire.Bind(new(forgotpassword.IdentityService), new(facade.IdentityFacade)),
	),

	wire.NewSet(
		user.DependencySet,
		wire.Bind(new(session.UserQuery), new(*user.Queries)),
		wire.Bind(new(interaction.UserService), new(*user.Provider)),
		wire.Bind(new(oidc.UserProvider), new(*user.Queries)),
		wire.Bind(new(event.UserService), new(*user.RawProvider)),
		wire.Bind(new(facade.UserCommands), new(*user.RawCommands)),
		wire.Bind(new(facade.UserProvider), new(*user.Provider)),
		wire.Bind(new(oauthhandler.TokenHandlerUserFacade), new(*user.Queries)),
	),

	wire.NewSet(
		sso.DependencySet,
		wire.Bind(new(interaction.OAuthProviderFactory), new(*sso.OAuthProviderFactory)),
	),

	wire.NewSet(
		welcomemessage.DependencySet,
		wire.Bind(new(user.WelcomeMessageProvider), new(*welcomemessage.Provider)),
	),

	wire.NewSet(
		forgotpassword.DependencySet,
		wire.Bind(new(interaction.ForgotPasswordService), new(*forgotpassword.Provider)),
		wire.Bind(new(interaction.ResetPasswordService), new(*forgotpassword.Provider)),
	),

	wire.NewSet(
		oauthpq.DependencySet,
		wire.Bind(new(oauth.AuthorizationStore), new(*oauthpq.AuthorizationStore)),
		wire.Bind(new(facade.OAuthService), new(*oauthpq.AuthorizationStore)),

		oauthredis.DependencySet,
		wire.Bind(new(oauth.AccessGrantStore), new(*oauthredis.Store)),
		wire.Bind(new(oauth.CodeGrantStore), new(*oauthredis.Store)),
		wire.Bind(new(oauth.OfflineGrantStore), new(*oauthredis.Store)),
		wire.Bind(new(oauth.AppSessionTokenStore), new(*oauthredis.Store)),
		wire.Bind(new(oauth.AppSessionStore), new(*oauthredis.Store)),

		oauth.DependencySet,
		wire.Bind(new(session.AccessTokenSessionResolver), new(*oauth.Resolver)),
		wire.Bind(new(session.AccessTokenSessionManager), new(*oauth.SessionManager)),
		wire.Bind(new(facade.OAuthSessionManager), new(*oauth.SessionManager)),
		wire.Bind(new(oauthhandler.OAuthURLProvider), new(*oauth.URLProvider)),
		wire.Value(oauthhandler.TokenGenerator(oauth.GenerateToken)),

		oauthhandler.DependencySet,

		oidc.DependencySet,
		wire.Value(oauthhandler.ScopesValidator(oidc.ValidateScopes)),
		wire.Bind(new(oauthhandler.IDTokenVerifier), new(*oidc.IDTokenIssuer)),
		wire.Bind(new(oauthhandler.IDTokenIssuer), new(*oidc.IDTokenIssuer)),
		wire.Bind(new(oauthhandler.AccessTokenIssuer), new(*oauth.AccessTokenEncoding)),
		wire.Bind(new(oauth.UserClaimsProvider), new(*oidc.IDTokenIssuer)),

		oidchandler.DependencySet,
	),

	wire.NewSet(
		interaction.DependencySet,
		wire.Bind(new(oauthhandler.GraphService), new(*interaction.Service)),
	),

	wire.NewSet(
		verification.DependencySet,
		wire.Bind(new(user.VerificationService), new(*verification.Service)),
		wire.Bind(new(facade.VerificationService), new(*verification.Service)),
		wire.Bind(new(interaction.VerificationService), new(*verification.Service)),
		wire.Bind(new(interaction.VerificationCodeSender), new(*verification.CodeSender)),
	),

	wire.NewSet(
		otp.DependencySet,
		wire.Bind(new(authenticatoroob.OTPMessageSender), new(*otp.MessageSender)),
		wire.Bind(new(verification.OTPMessageSender), new(*otp.MessageSender)),
	),

	wire.NewSet(
		translation.DependencySet,
		wire.Bind(new(otp.TranslationService), new(*translation.Service)),
		wire.Bind(new(forgotpassword.TranslationService), new(*translation.Service)),
		wire.Bind(new(welcomemessage.TranslationService), new(*translation.Service)),
	),

	wire.NewSet(
		web.DependencySet,
		wire.Bind(new(translation.StaticAssetResolver), new(*web.StaticAssetResolver)),
	),

	wire.NewSet(
		ratelimit.DependencySet,
		wire.Bind(new(interaction.RateLimiter), new(*ratelimit.Limiter)),
		wire.Bind(new(authenticatorservice.RateLimiter), new(*ratelimit.Limiter)),
		wire.Bind(new(otp.RateLimiter), new(*ratelimit.Limiter)),
		wire.Bind(new(forgotpassword.RateLimiter), new(*ratelimit.Limiter)),
		wire.Bind(new(welcomemessage.RateLimiter), new(*ratelimit.Limiter)),
		wire.Bind(new(mfa.RateLimiter), new(*ratelimit.Limiter)),
		wire.Bind(new(verification.RateLimiter), new(*ratelimit.Limiter)),
	),
)
