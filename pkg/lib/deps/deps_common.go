package deps

import (
	"github.com/google/wire"

	authenticatoroob "github.com/authgear/authgear-server/pkg/lib/authn/authenticator/oob"
	authenticatorpassword "github.com/authgear/authgear-server/pkg/lib/authn/authenticator/password"
	authenticatorservice "github.com/authgear/authgear-server/pkg/lib/authn/authenticator/service"
	authenticatortotp "github.com/authgear/authgear-server/pkg/lib/authn/authenticator/totp"
	"github.com/authgear/authgear-server/pkg/lib/authn/challenge"
	identityanonymous "github.com/authgear/authgear-server/pkg/lib/authn/identity/anonymous"
	identityloginid "github.com/authgear/authgear-server/pkg/lib/authn/identity/loginid"
	identityoauth "github.com/authgear/authgear-server/pkg/lib/authn/identity/oauth"
	identityservice "github.com/authgear/authgear-server/pkg/lib/authn/identity/service"
	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/authn/sso"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/facade"
	"github.com/authgear/authgear-server/pkg/lib/feature/forgotpassword"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/feature/welcomemessage"
	"github.com/authgear/authgear-server/pkg/lib/hook"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	oauthhandler "github.com/authgear/authgear-server/pkg/lib/oauth/handler"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	oidchandler "github.com/authgear/authgear-server/pkg/lib/oauth/oidc/handler"
	oauthpq "github.com/authgear/authgear-server/pkg/lib/oauth/pq"
	oauthredis "github.com/authgear/authgear-server/pkg/lib/oauth/redis"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/access"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/lib/translation"
)

var CommonDependencySet = wire.NewSet(
	configDeps,
	utilsDeps,

	db.DependencySet,

	wire.NewSet(
		challenge.DependencySet,
		wire.Bind(new(interaction.ChallengeProvider), new(*challenge.Provider)),
	),

	wire.NewSet(
		hook.DependencySet,
		wire.Bind(new(interaction.HookProvider), new(*hook.Provider)),
		wire.Bind(new(user.HookProvider), new(*hook.Provider)),
		wire.Bind(new(session.HookProvider), new(*hook.Provider)),
	),

	wire.NewSet(
		idpsession.DependencySet,

		wire.Bind(new(session.IDPSessionResolver), new(*idpsession.Resolver)),
		wire.Bind(new(session.IDPSessionManager), new(*idpsession.Manager)),
		wire.Bind(new(oauth.ResolverSessionProvider), new(*idpsession.Provider)),
		wire.Bind(new(oauthhandler.SessionProvider), new(*idpsession.Provider)),
		wire.Bind(new(interaction.SessionProvider), new(*idpsession.Provider)),
	),

	wire.NewSet(
		access.DependencySet,
		session.DependencySet,
		wire.Bind(new(idpsession.AccessEventProvider), new(*access.EventProvider)),
	),

	wire.NewSet(
		authenticatorpassword.DependencySet,
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

		identityservice.DependencySet,
		wire.Bind(new(identityservice.LoginIDIdentityProvider), new(*identityloginid.Provider)),
		wire.Bind(new(identityservice.OAuthIdentityProvider), new(*identityoauth.Provider)),
		wire.Bind(new(identityservice.AnonymousIdentityProvider), new(*identityanonymous.Provider)),

		wire.Bind(new(facade.IdentityService), new(*identityservice.Service)),
	),

	wire.NewSet(
		facade.DependencySet,

		wire.Bind(new(interaction.IdentityService), new(facade.IdentityFacade)),
		wire.Bind(new(interaction.AuthenticatorService), new(facade.AuthenticatorFacade)),
		wire.Bind(new(forgotpassword.AuthenticatorService), new(facade.AuthenticatorFacade)),
		wire.Bind(new(user.IdentityService), new(facade.IdentityFacade)),
		wire.Bind(new(forgotpassword.IdentityService), new(facade.IdentityFacade)),
	),

	wire.NewSet(
		user.DependencySet,
		wire.Bind(new(session.UserProvider), new(*user.Queries)),
		wire.Bind(new(interaction.UserService), new(*user.Provider)),
		wire.Bind(new(oidc.UserProvider), new(*user.Queries)),
		wire.Bind(new(hook.UserProvider), new(*user.RawProvider)),
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

		oauthredis.DependencySet,
		wire.Bind(new(oauth.AccessGrantStore), new(*oauthredis.GrantStore)),
		wire.Bind(new(oauth.CodeGrantStore), new(*oauthredis.GrantStore)),
		wire.Bind(new(oauth.OfflineGrantStore), new(*oauthredis.GrantStore)),

		oauth.DependencySet,
		wire.Bind(new(session.AccessTokenSessionResolver), new(*oauth.Resolver)),
		wire.Bind(new(session.AccessTokenSessionManager), new(*oauth.SessionManager)),
		wire.Bind(new(oauthhandler.OAuthURLProvider), new(*oauth.URLProvider)),
		wire.Value(oauthhandler.TokenGenerator(oauth.GenerateToken)),

		oauthhandler.DependencySet,
	),

	wire.NewSet(
		oidc.DependencySet,
		wire.Value(oauthhandler.ScopesValidator(oidc.ValidateScopes)),
		wire.Bind(new(oauthhandler.IDTokenIssuer), new(*oidc.IDTokenIssuer)),

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
)
