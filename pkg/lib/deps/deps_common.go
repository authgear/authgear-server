package deps

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/lib/audit"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	authenticatoroob "github.com/authgear/authgear-server/pkg/lib/authn/authenticator/oob"
	authenticatorpasskey "github.com/authgear/authgear-server/pkg/lib/authn/authenticator/passkey"
	authenticatorpassword "github.com/authgear/authgear-server/pkg/lib/authn/authenticator/password"
	authenticatorservice "github.com/authgear/authgear-server/pkg/lib/authn/authenticator/service"
	authenticatortotp "github.com/authgear/authgear-server/pkg/lib/authn/authenticator/totp"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/whatsapp"
	"github.com/authgear/authgear-server/pkg/lib/authn/challenge"
	identityanonymous "github.com/authgear/authgear-server/pkg/lib/authn/identity/anonymous"
	identitybiometric "github.com/authgear/authgear-server/pkg/lib/authn/identity/biometric"
	identityloginid "github.com/authgear/authgear-server/pkg/lib/authn/identity/loginid"
	identityoauth "github.com/authgear/authgear-server/pkg/lib/authn/identity/oauth"
	identitypasskey "github.com/authgear/authgear-server/pkg/lib/authn/identity/passkey"
	identityservice "github.com/authgear/authgear-server/pkg/lib/authn/identity/service"
	identitysiwe "github.com/authgear/authgear-server/pkg/lib/authn/identity/siwe"
	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/authn/sso"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	libes "github.com/authgear/authgear-server/pkg/lib/elasticsearch"
	"github.com/authgear/authgear-server/pkg/lib/event"
	"github.com/authgear/authgear-server/pkg/lib/facade"
	featurecustomattrs "github.com/authgear/authgear-server/pkg/lib/feature/customattrs"
	"github.com/authgear/authgear-server/pkg/lib/feature/forgotpassword"
	featurepasskey "github.com/authgear/authgear-server/pkg/lib/feature/passkey"
	featuresiwe "github.com/authgear/authgear-server/pkg/lib/feature/siwe"
	featurestdattrs "github.com/authgear/authgear-server/pkg/lib/feature/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	featureweb3 "github.com/authgear/authgear-server/pkg/lib/feature/web3"
	"github.com/authgear/authgear-server/pkg/lib/feature/welcomemessage"
	"github.com/authgear/authgear-server/pkg/lib/healthz"
	"github.com/authgear/authgear-server/pkg/lib/hook"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/meter"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	oauthhandler "github.com/authgear/authgear-server/pkg/lib/oauth/handler"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oauthsession"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	oidchandler "github.com/authgear/authgear-server/pkg/lib/oauth/oidc/handler"
	oauthpq "github.com/authgear/authgear-server/pkg/lib/oauth/pq"
	oauthredis "github.com/authgear/authgear-server/pkg/lib/oauth/redis"
	"github.com/authgear/authgear-server/pkg/lib/presign"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/access"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/lib/sessionlisting"
	"github.com/authgear/authgear-server/pkg/lib/translation"
	"github.com/authgear/authgear-server/pkg/lib/tutorial"
	"github.com/authgear/authgear-server/pkg/lib/usage"
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
		oauthsession.DependencySet,
		wire.Bind(new(oauthhandler.OAuthSessionService), new(*oauthsession.StoreRedis)),
	),

	wire.NewSet(
		libes.DependencySet,
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
		wire.Bind(new(welcomemessage.EventService), new(*event.Service)),
		wire.Bind(new(featurestdattrs.EventService), new(*event.Service)),
		wire.Bind(new(featurecustomattrs.EventService), new(*event.Service)),
		wire.Bind(new(facade.EventService), new(*event.Service)),
		wire.Bind(new(whatsapp.EventService), new(*event.Service)),
		wire.Bind(new(oauthhandler.EventService), new(*event.Service)),
		wire.Bind(new(oauth.EventService), new(*event.Service)),
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
		wire.Bind(new(oauth.ServiceIDPSessionProvider), new(*idpsession.Provider)),
		wire.Bind(new(oauthhandler.SessionProvider), new(*idpsession.Provider)),
		wire.Bind(new(interaction.SessionProvider), new(*idpsession.Provider)),
		wire.Bind(new(sessionlisting.IDPSessionProvider), new(*idpsession.Provider)),
		wire.Bind(new(facade.IDPSessionManager), new(*idpsession.Manager)),
	),

	wire.NewSet(
		access.DependencySet,
		session.DependencySet,
		wire.Bind(new(idpsession.AccessEventProvider), new(*access.EventProvider)),
		wire.Bind(new(oidchandler.LogoutSessionManager), new(*session.Manager)),
		wire.Bind(new(oauthhandler.SessionManager), new(*session.Manager)),
		wire.Bind(new(interaction.SessionManager), new(*session.Manager)),
	),

	wire.NewSet(
		authenticatorpassword.DependencySet,
		wire.Bind(new(facade.PasswordHistoryStore), new(*authenticatorpassword.HistoryStore)),
		authenticatoroob.DependencySet,
		wire.Bind(new(interaction.OOBCodeSender), new(*authenticatoroob.CodeSender)),
		authenticatortotp.DependencySet,
		authenticatorpasskey.DependencySet,

		authenticatorservice.DependencySet,
		wire.Bind(new(authenticatorservice.PasswordAuthenticatorProvider), new(*authenticatorpassword.Provider)),
		wire.Bind(new(authenticatorservice.PasskeyAuthenticatorProvider), new(*authenticatorpasskey.Provider)),
		wire.Bind(new(authenticatorservice.OOBOTPAuthenticatorProvider), new(*authenticatoroob.Provider)),
		wire.Bind(new(authenticatorservice.TOTPAuthenticatorProvider), new(*authenticatortotp.Provider)),

		wire.Bind(new(facade.AuthenticatorService), new(*authenticatorservice.Service)),
		wire.Bind(new(user.AuthenticatorService), new(*authenticatorservice.Service)),

		whatsapp.DependencySet,
		wire.Bind(new(interaction.WhatsappCodeProvider), new(*whatsapp.Provider)),
	),

	wire.NewSet(
		mfa.DependencySet,

		wire.Bind(new(interaction.MFAService), new(*mfa.Service)),
		wire.Bind(new(facade.MFAService), new(*mfa.Service)),
	),

	wire.NewSet(
		stdattrs.DependencySet,
		wire.Bind(new(sso.StandardAttributesNormalizer), new(*stdattrs.Normalizer)),
	),

	wire.NewSet(
		featurecustomattrs.DependencySet,
		wire.Bind(new(user.CustomAttributesService), new(*featurecustomattrs.ServiceNoEvent)),
		wire.Bind(new(hook.CustomAttributesServiceNoEvent), new(*featurecustomattrs.ServiceNoEvent)),
	),

	wire.NewSet(
		identityloginid.DependencySet,
		wire.Bind(new(stdattrs.LoginIDNormalizerFactory), new(*identityloginid.NormalizerFactory)),
		wire.Bind(new(interaction.LoginIDNormalizerFactory), new(*identityloginid.NormalizerFactory)),

		identityoauth.DependencySet,

		identityanonymous.DependencySet,
		wire.Bind(new(interaction.AnonymousIdentityProvider), new(*identityanonymous.Provider)),
		wire.Bind(new(interaction.AnonymousUserPromotionCodeStore), new(*identityanonymous.StoreRedis)),

		identitypasskey.DependencySet,

		identitybiometric.DependencySet,
		wire.Bind(new(interaction.BiometricIdentityProvider), new(*identitybiometric.Provider)),

		identitysiwe.DependencySet,

		identityservice.DependencySet,
		wire.Bind(new(identityservice.LoginIDIdentityProvider), new(*identityloginid.Provider)),
		wire.Bind(new(identityservice.OAuthIdentityProvider), new(*identityoauth.Provider)),
		wire.Bind(new(identityservice.PasskeyIdentityProvider), new(*identitypasskey.Provider)),
		wire.Bind(new(identityservice.AnonymousIdentityProvider), new(*identityanonymous.Provider)),
		wire.Bind(new(identityservice.BiometricIdentityProvider), new(*identitybiometric.Provider)),
		wire.Bind(new(identityservice.SIWEIdentityProvider), new(*identitysiwe.Provider)),

		wire.Bind(new(facade.IdentityService), new(*identityservice.Service)),
		wire.Bind(new(user.IdentityService), new(*identityservice.Service)),
		wire.Bind(new(featurestdattrs.IdentityService), new(*identityservice.Service)),
		wire.Bind(new(featurepasskey.IdentityService), new(*identityservice.Service)),

		wire.Bind(new(oauthhandler.PromotionCodeStore), new(*identityanonymous.StoreRedis)),
		wire.Bind(new(oauthhandler.AnonymousIdentityProvider), new(*identityanonymous.Provider)),
	),

	wire.NewSet(
		facade.DependencySet,

		wire.Bind(new(interaction.IdentityService), new(facade.IdentityFacade)),
		wire.Bind(new(interaction.AuthenticatorService), new(facade.AuthenticatorFacade)),
		wire.Bind(new(forgotpassword.AuthenticatorService), new(facade.AuthenticatorFacade)),
		wire.Bind(new(forgotpassword.IdentityService), new(facade.IdentityFacade)),
	),

	wire.NewSet(
		user.DependencySet,
		wire.Bind(new(session.UserQuery), new(*user.Queries)),
		wire.Bind(new(interaction.UserService), new(*user.Provider)),
		wire.Bind(new(oidc.UserProvider), new(*user.Queries)),
		wire.Bind(new(featurestdattrs.UserQueries), new(*user.RawQueries)),
		wire.Bind(new(featurestdattrs.UserStore), new(*user.Store)),
		wire.Bind(new(featurecustomattrs.UserStore), new(*user.Store)),
		wire.Bind(new(featurecustomattrs.UserQueries), new(*user.RawQueries)),
		wire.Bind(new(featurepasskey.UserService), new(*user.Queries)),
		wire.Bind(new(facade.UserCommands), new(*user.Commands)),
		wire.Bind(new(facade.UserQueries), new(*user.Queries)),
		wire.Bind(new(facade.UserProvider), new(*user.Provider)),
		wire.Bind(new(oauthhandler.TokenHandlerUserFacade), new(*user.Queries)),
		wire.Bind(new(oauthhandler.UserProvider), new(*user.Queries)),
		wire.Bind(new(event.ResolverUserQueries), new(*user.Queries)),
		wire.Bind(new(libes.UserQueries), new(*user.Queries)),
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
		wire.Bind(new(oauthhandler.AppSessionTokenService), new(*oauth.AppSessionTokenService)),
		wire.Bind(new(sessionlisting.OfflineGrantService), new(*oauth.OfflineGrantService)),
		wire.Value(oauthhandler.TokenGenerator(oauth.GenerateToken)),
		wire.Bind(new(oauthhandler.AuthorizationService), new(*oauth.AuthorizationService)),
		wire.Bind(new(interaction.OfflineGrantStore), new(*oauthredis.Store)),

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
		wire.Bind(new(webapp.AntiSpamOTPCodeBucketMaker), new(*interaction.AntiSpamOTPCodeBucketMaker)),
	),

	wire.NewSet(
		verification.DependencySet,
		wire.Bind(new(featurestdattrs.ClaimStore), new(*verification.StorePQ)),
		wire.Bind(new(user.VerificationService), new(*verification.Service)),
		wire.Bind(new(facade.VerificationService), new(*verification.Service)),
		wire.Bind(new(interaction.VerificationService), new(*verification.Service)),
	),

	wire.NewSet(
		otp.DependencySet,
		wire.Bind(new(authenticatorservice.OTPCodeService), new(*otp.Service)),
		wire.Bind(new(interaction.OTPCodeService), new(*otp.Service)),
		wire.Bind(new(whatsapp.OTPCodeService), new(*otp.Service)),
		wire.Bind(new(webapp.OTPCodeService), new(*otp.Service)),
		wire.Bind(new(authenticatoroob.OTPMessageSender), new(*otp.MessageSender)),
	),

	wire.NewSet(
		translation.DependencySet,
		wire.Bind(new(otp.TranslationService), new(*translation.Service)),
		wire.Bind(new(forgotpassword.TranslationService), new(*translation.Service)),
		wire.Bind(new(welcomemessage.TranslationService), new(*translation.Service)),
		wire.Bind(new(featurepasskey.TranslationService), new(*translation.Service)),
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
		wire.Bind(new(featuresiwe.RateLimiter), new(*ratelimit.Limiter)),
	),

	wire.NewSet(
		meter.DependencySet,
		wire.Bind(new(session.MeterService), new(*meter.Service)),
	),

	wire.NewSet(
		featurestdattrs.DependencySet,
		wire.Bind(new(user.StandardAttributesService), new(*featurestdattrs.ServiceNoEvent)),
		wire.Bind(new(facade.StdAttrsService), new(*featurestdattrs.Service)),
		wire.Bind(new(interaction.StdAttrsService), new(*featurestdattrs.Service)),
		wire.Bind(new(hook.StandardAttributesServiceNoEvent), new(*featurestdattrs.ServiceNoEvent)),
	),

	presign.DependencySet,

	tutorial.DependencySet,

	sessionlisting.DependencySet,

	wire.NewSet(
		usage.DependencySet,
		wire.Bind(new(forgotpassword.HardSMSBucketer), new(*usage.HardSMSBucketer)),
		wire.Bind(new(otp.HardSMSBucketer), new(*usage.HardSMSBucketer)),
	),

	wire.NewSet(
		sms.DependencySet,
		wire.Bind(new(forgotpassword.AntiSpamSMSBucketMaker), new(*sms.AntiSpamSMSBucketMaker)),
		wire.Bind(new(otp.AntiSpamSMSBucketMaker), new(*sms.AntiSpamSMSBucketMaker)),
	),

	wire.NewSet(
		featurepasskey.DependencySet,
		wire.Bind(new(identitypasskey.PasskeyService), new(*featurepasskey.Service)),
		wire.Bind(new(authenticatorpasskey.PasskeyService), new(*featurepasskey.Service)),
		wire.Bind(new(interaction.PasskeyService), new(*featurepasskey.Service)),
	),

	wire.NewSet(
		featuresiwe.DependencySet,
		wire.Bind(new(identitysiwe.SIWEService), new(*featuresiwe.Service)),
	),

	wire.NewSet(
		featureweb3.DependencySet,
		wire.Bind(new(user.Web3Service), new(*featureweb3.Service)),
	),
)
