package deps

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	handlerwebappauthflowv2 "github.com/authgear/authgear-server/pkg/auth/handler/webapp/authflowv2"
	"github.com/authgear/authgear-server/pkg/lib/accountmanagement"
	"github.com/authgear/authgear-server/pkg/lib/accountmigration"
	"github.com/authgear/authgear-server/pkg/lib/app2app"
	"github.com/authgear/authgear-server/pkg/lib/audit"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	authenticatoroob "github.com/authgear/authgear-server/pkg/lib/authn/authenticator/oob"
	authenticatorpasskey "github.com/authgear/authgear-server/pkg/lib/authn/authenticator/passkey"
	authenticatorpassword "github.com/authgear/authgear-server/pkg/lib/authn/authenticator/password"
	authenticatorservice "github.com/authgear/authgear-server/pkg/lib/authn/authenticator/service"
	authenticatortotp "github.com/authgear/authgear-server/pkg/lib/authn/authenticator/totp"
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
	"github.com/authgear/authgear-server/pkg/lib/botprotection"
	"github.com/authgear/authgear-server/pkg/lib/dpop"
	libes "github.com/authgear/authgear-server/pkg/lib/elasticsearch"
	"github.com/authgear/authgear-server/pkg/lib/endpoints"
	"github.com/authgear/authgear-server/pkg/lib/event"
	"github.com/authgear/authgear-server/pkg/lib/facade"
	"github.com/authgear/authgear-server/pkg/lib/feature/captcha"
	featurecustomattrs "github.com/authgear/authgear-server/pkg/lib/feature/customattrs"
	"github.com/authgear/authgear-server/pkg/lib/feature/forgotpassword"
	featurepasskey "github.com/authgear/authgear-server/pkg/lib/feature/passkey"
	featuresiwe "github.com/authgear/authgear-server/pkg/lib/feature/siwe"
	featurestdattrs "github.com/authgear/authgear-server/pkg/lib/feature/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	featureweb3 "github.com/authgear/authgear-server/pkg/lib/feature/web3"
	"github.com/authgear/authgear-server/pkg/lib/healthz"
	"github.com/authgear/authgear-server/pkg/lib/hook"

	deprecated_infracaptcha "github.com/authgear/authgear-server/pkg/lib/infra/captcha"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/redisqueue"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms"
	infrawhatsapp "github.com/authgear/authgear-server/pkg/lib/infra/whatsapp"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/lockout"
	"github.com/authgear/authgear-server/pkg/lib/messaging"
	"github.com/authgear/authgear-server/pkg/lib/meter"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/handler"
	oauthhandler "github.com/authgear/authgear-server/pkg/lib/oauth/handler"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oauthsession"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	oidchandler "github.com/authgear/authgear-server/pkg/lib/oauth/oidc/handler"
	oauthpq "github.com/authgear/authgear-server/pkg/lib/oauth/pq"
	oauthredis "github.com/authgear/authgear-server/pkg/lib/oauth/redis"
	"github.com/authgear/authgear-server/pkg/lib/oauthclient"
	"github.com/authgear/authgear-server/pkg/lib/presign"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/rolesgroups"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/access"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/lib/sessionlisting"
	"github.com/authgear/authgear-server/pkg/lib/tester"
	"github.com/authgear/authgear-server/pkg/lib/translation"
	"github.com/authgear/authgear-server/pkg/lib/tutorial"
	"github.com/authgear/authgear-server/pkg/lib/usage"
	"github.com/authgear/authgear-server/pkg/lib/userimport"
	"github.com/authgear/authgear-server/pkg/lib/web"
	"github.com/authgear/authgear-server/pkg/lib/webappoauth"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var CommonDependencySet = wire.NewSet(
	ConfigDeps,
	utilsDeps,

	appdb.DependencySet,
	auditdb.DependencySet,
	template.DependencySet,

	healthz.DependencySet,

	wire.NewSet(
		authenticationinfo.DependencySet,
		wire.Bind(new(interaction.AuthenticationInfoService), new(*authenticationinfo.StoreRedis)),
		wire.Bind(new(workflow.AuthenticationInfoService), new(*authenticationinfo.StoreRedis)),
		wire.Bind(new(authenticationflow.AuthenticationInfoService), new(*authenticationinfo.StoreRedis)),
		wire.Bind(new(oauthhandler.AuthenticationInfoService), new(*authenticationinfo.StoreRedis)),
	),

	wire.NewSet(
		oauthsession.DependencySet,
		wire.Bind(new(oauthhandler.OAuthSessionService), new(*oauthsession.StoreRedis)),
		wire.Bind(new(interaction.OAuthSessions), new(*oauthsession.StoreRedis)),
	),

	wire.NewSet(
		libes.DependencySet,
		wire.Bind(new(userimport.ElasticsearchService), new(*libes.Service)),
	),

	wire.NewSet(
		challenge.DependencySet,
		wire.Bind(new(interaction.ChallengeProvider), new(*challenge.Provider)),
		wire.Bind(new(oauthhandler.ChallengeProvider), new(*challenge.Provider)),
		wire.Bind(new(authenticationflow.ChallengeService), new(*challenge.Provider)),
	),

	wire.NewSet(
		event.DependencySet,
		wire.Bind(new(interaction.EventService), new(*event.Service)),
		wire.Bind(new(workflow.EventService), new(*event.Service)),
		wire.Bind(new(authenticationflow.EventService), new(*event.Service)),
		wire.Bind(new(accountmanagement.EventService), new(*event.Service)),
		wire.Bind(new(user.EventService), new(*event.Service)),
		wire.Bind(new(session.EventService), new(*event.Service)),
		wire.Bind(new(messaging.EventService), new(*event.Service)),
		wire.Bind(new(featurestdattrs.EventService), new(*event.Service)),
		wire.Bind(new(featurecustomattrs.EventService), new(*event.Service)),
		wire.Bind(new(facade.EventService), new(*event.Service)),
		wire.Bind(new(oauthhandler.EventService), new(*event.Service)),
		wire.Bind(new(oauth.EventService), new(*event.Service)),
		wire.Bind(new(botprotection.EventService), new(*event.Service)),
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
		wire.Bind(new(oidc.IDTokenHintResolverSessionProvider), new(*idpsession.Provider)),
		wire.Bind(new(interaction.SessionProvider), new(*idpsession.Provider)),
		wire.Bind(new(workflow.IDPSessionService), new(*idpsession.Provider)),
		wire.Bind(new(oauth.IDPSessionProvider), new(*idpsession.Provider)),
		wire.Bind(new(authenticationflow.IDPSessionService), new(*idpsession.Provider)),
		wire.Bind(new(sessionlisting.IDPSessionProvider), new(*idpsession.Provider)),
		wire.Bind(new(facade.IDPSessionManager), new(*idpsession.Manager)),
	),

	wire.NewSet(
		dpop.DependencySet,
	),

	wire.NewSet(
		access.DependencySet,
		session.DependencySet,
		wire.Bind(new(idpsession.AccessEventProvider), new(*access.EventProvider)),
		wire.Bind(new(oidchandler.LogoutSessionManager), new(*session.Manager)),
		wire.Bind(new(oauthhandler.SessionManager), new(*session.Manager)),
		wire.Bind(new(interaction.SessionManager), new(*session.Manager)),
		wire.Bind(new(workflow.SessionService), new(*session.Manager)),
		wire.Bind(new(authenticationflow.SessionService), new(*session.Manager)),
	),

	wire.NewSet(
		authenticatorpassword.DependencySet,
		wire.Bind(new(facade.PasswordHistoryStore), new(*authenticatorpassword.HistoryStore)),
		authenticatoroob.DependencySet,
		authenticatortotp.DependencySet,
		authenticatorpasskey.DependencySet,

		authenticatorservice.DependencySet,
		wire.Bind(new(authenticatorservice.PasswordAuthenticatorProvider), new(*authenticatorpassword.Provider)),
		wire.Bind(new(authenticatorservice.PasskeyAuthenticatorProvider), new(*authenticatorpasskey.Provider)),
		wire.Bind(new(authenticatorservice.OOBOTPAuthenticatorProvider), new(*authenticatoroob.Provider)),
		wire.Bind(new(authenticatorservice.TOTPAuthenticatorProvider), new(*authenticatortotp.Provider)),

		wire.Bind(new(facade.AuthenticatorService), new(*authenticatorservice.Service)),
		wire.Bind(new(user.AuthenticatorService), new(*authenticatorservice.Service)),
	),

	wire.NewSet(
		mfa.DependencySet,

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
		wire.Bind(new(workflow.CustomAttrsService), new(*featurecustomattrs.Service)),
		wire.Bind(new(authenticationflow.CustomAttrsService), new(*featurecustomattrs.Service)),
		wire.Bind(new(userimport.CustomAttributesService), new(*featurecustomattrs.ServiceNoEvent)),
	),

	wire.NewSet(
		identityloginid.DependencySet,
		wire.Bind(new(stdattrs.LoginIDNormalizerFactory), new(*identityloginid.NormalizerFactory)),
		wire.Bind(new(authenticatoroob.LoginIDNormalizerFactory), new(*identityloginid.NormalizerFactory)),
		wire.Bind(new(authenticationflow.LoginIDService), new(*identityloginid.Provider)),

		identityoauth.DependencySet,

		identityanonymous.DependencySet,
		wire.Bind(new(interaction.AnonymousIdentityProvider), new(*identityanonymous.Provider)),
		wire.Bind(new(interaction.AnonymousUserPromotionCodeStore), new(*identityanonymous.StoreRedis)),
		wire.Bind(new(authenticationflow.AnonymousIdentityService), new(*identityanonymous.Provider)),
		wire.Bind(new(authenticationflow.AnonymousUserPromotionCodeStore), new(*identityanonymous.StoreRedis)),

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
		wire.Bind(new(workflow.AuthenticatorService), new(facade.AuthenticatorFacade)),
		wire.Bind(new(authenticationflow.AuthenticatorService), new(facade.AuthenticatorFacade)),
		wire.Bind(new(forgotpassword.AuthenticatorService), new(facade.AuthenticatorFacade)),
		wire.Bind(new(forgotpassword.IdentityService), new(facade.IdentityFacade)),
		wire.Bind(new(workflow.IdentityService), new(facade.IdentityFacade)),
		wire.Bind(new(workflow.VerificationService), new(facade.WorkflowVerificationFacade)),
		wire.Bind(new(workflow.MFAService), new(*facade.MFAFacade)),
		wire.Bind(new(authenticationflow.IdentityService), new(facade.IdentityFacade)),
		wire.Bind(new(authenticationflow.VerificationService), new(facade.WorkflowVerificationFacade)),
		wire.Bind(new(authenticationflow.MFAService), new(*facade.MFAFacade)),
		wire.Bind(new(interaction.MFAService), new(*facade.MFAFacade)),
		wire.Bind(new(userimport.IdentityService), new(*facade.IdentityFacade)),
		wire.Bind(new(userimport.AuthenticatorService), new(*facade.AuthenticatorFacade)),
		wire.Bind(new(accountmanagement.IdentityService), new(*facade.IdentityFacade)),
	),

	wire.NewSet(
		user.DependencySet,
		wire.Bind(new(session.UserQuery), new(*user.Queries)),
		wire.Bind(new(interaction.UserService), new(*user.Provider)),
		wire.Bind(new(workflow.UserService), new(*user.Provider)),
		wire.Bind(new(authenticationflow.UserService), new(*user.Provider)),
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
		wire.Bind(new(userimport.UserCommands), new(*user.RawCommands)),
		wire.Bind(new(userimport.UserQueries), new(*user.RawQueries)),
	),

	wire.NewSet(
		rolesgroups.DependencySet,
		wire.Bind(new(facade.RolesGroupsCommands), new(*rolesgroups.Commands)),
		wire.Bind(new(oidc.RolesAndGroupsProvider), new(*rolesgroups.Queries)),
		wire.Bind(new(hook.RolesAndGroupsServiceNoEvent), new(*rolesgroups.Commands)),
		wire.Bind(new(user.RolesAndGroupsService), new(*rolesgroups.Queries)),
		wire.Bind(new(userimport.RolesGroupsCommands), new(*rolesgroups.Commands)),
	),

	wire.NewSet(
		sso.DependencySet,
		wire.Bind(new(interaction.OAuthProviderFactory), new(*sso.OAuthProviderFactory)),
		wire.Bind(new(authenticationflow.OAuthProviderFactory), new(*sso.OAuthProviderFactory)),
		wire.Bind(new(accountmanagement.OAuthProvider), new(*sso.OAuthProviderFactory)),
	),

	wire.NewSet(
		forgotpassword.DependencySet,
		wire.Bind(new(interaction.ForgotPasswordService), new(*forgotpassword.Service)),
		wire.Bind(new(interaction.ResetPasswordService), new(*forgotpassword.Service)),
		wire.Bind(new(workflow.ForgotPasswordService), new(*forgotpassword.Service)),
		wire.Bind(new(authenticationflow.ForgotPasswordService), new(*forgotpassword.Service)),
		wire.Bind(new(workflow.ResetPasswordService), new(*forgotpassword.Service)),
		wire.Bind(new(authenticationflow.ResetPasswordService), new(*forgotpassword.Service)),
	),

	wire.NewSet(
		captcha.DependencySet,
		wire.Bind(new(workflow.CaptchaService), new(*captcha.Provider)),
		wire.Bind(new(authenticationflow.CaptchaService), new(*captcha.Provider)),
	),

	wire.NewSet(
		botprotection.DependencySet,
		wire.Bind(new(authenticationflow.BotProtectionService), new(*botprotection.Provider)),
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
		wire.Bind(new(oauth.PreAuthenticatedURLTokenStore), new(*oauthredis.Store)),
		wire.Bind(new(oauth.SettingsActionGrantStore), new(*oauthredis.Store)),
		wire.Bind(new(handler.TokenHandlerCodeGrantStore), new(*oauthredis.Store)),
		wire.Bind(new(handler.TokenHandlerSettingsActionGrantStore), new(*oauthredis.Store)),
		wire.Bind(new(handler.TokenHandlerOfflineGrantStore), new(*oauthredis.Store)),
		wire.Bind(new(handler.TokenHandlerAppSessionTokenStore), new(*oauthredis.Store)),

		oauth.DependencySet,
		wire.Bind(new(session.AccessTokenSessionResolver), new(*oauth.Resolver)),
		wire.Bind(new(session.AccessTokenSessionManager), new(*oauth.SessionManager)),
		wire.Bind(new(facade.OAuthSessionManager), new(*oauth.SessionManager)),
		wire.Bind(new(oauthhandler.AppSessionTokenService), new(*oauth.AppSessionTokenService)),
		wire.Bind(new(sessionlisting.OfflineGrantService), new(*oauth.OfflineGrantService)),
		wire.Bind(new(handler.TokenHandlerOfflineGrantService), new(*oauth.OfflineGrantService)),
		wire.Value(oauthhandler.TokenGenerator(oauth.GenerateToken)),
		wire.Bind(new(oauthhandler.AuthorizationService), new(*oauth.AuthorizationService)),
		wire.Bind(new(interaction.OfflineGrantStore), new(*oauthredis.Store)),
		wire.Bind(new(workflow.OfflineGrantStore), new(*oauthredis.Store)),
		wire.Bind(new(authenticationflow.OfflineGrantStore), new(*oauthredis.Store)),
		wire.Bind(new(oidc.UIInfoResolverPromptResolver), new(*oauth.PromptResolver)),

		oauthhandler.DependencySet,

		oidc.DependencySet,
		wire.Value(oauthhandler.ScopesValidator(oidc.ValidateScopes)),
		wire.Bind(new(oauthhandler.UIInfoResolver), new(*oidc.UIInfoResolver)),
		wire.Bind(new(webapp.SelectAccountUIInfoResolver), new(*oidc.UIInfoResolver)),
		wire.Bind(new(handlerwebappauthflowv2.SelectAccountUIInfoResolver), new(*oidc.UIInfoResolver)),
		wire.Bind(new(workflow.ServiceUIInfoResolver), new(*oidc.UIInfoResolver)),
		wire.Bind(new(authenticationflow.ServiceUIInfoResolver), new(*oidc.UIInfoResolver)),
		wire.Bind(new(authenticationflow.IDTokenService), new(*oidc.IDTokenIssuer)),
		wire.Bind(new(oauthhandler.IDTokenIssuer), new(*oidc.IDTokenIssuer)),
		wire.Bind(new(oauthhandler.AccessTokenIssuer), new(*oauth.AccessTokenEncoding)),
		wire.Bind(new(oauth.UserClaimsProvider), new(*oidc.IDTokenIssuer)),
		wire.Bind(new(oauthhandler.UIURLBuilder), new(*oidc.UIURLBuilder)),

		oidchandler.DependencySet,
	),

	wire.NewSet(
		interaction.DependencySet,
		wire.Bind(new(oauthhandler.GraphService), new(*interaction.Service)),
	),

	wire.NewSet(
		verification.DependencySet,
		wire.Bind(new(featurestdattrs.ClaimStore), new(*verification.StorePQ)),
		wire.Bind(new(user.VerificationService), new(*verification.Service)),
		wire.Bind(new(facade.VerificationService), new(*verification.Service)),
		wire.Bind(new(interaction.VerificationService), new(*verification.Service)),
		wire.Bind(new(userimport.VerifiedClaimService), new(*verification.Service)),
	),

	wire.NewSet(
		otp.DependencySet,
		wire.Bind(new(authenticatorservice.OTPCodeService), new(*otp.Service)),
		wire.Bind(new(interaction.OTPCodeService), new(*otp.Service)),
		wire.Bind(new(workflow.OTPCodeService), new(*otp.Service)),
		wire.Bind(new(authenticationflow.OTPCodeService), new(*otp.Service)),
		wire.Bind(new(webapp.OTPCodeService), new(*otp.Service)),
		wire.Bind(new(forgotpassword.OTPCodeService), new(*otp.Service)),
		wire.Bind(new(interaction.OTPSender), new(*otp.MessageSender)),
		wire.Bind(new(workflow.OTPSender), new(*otp.MessageSender)),
		wire.Bind(new(authenticationflow.OTPSender), new(*otp.MessageSender)),
		wire.Bind(new(forgotpassword.OTPSender), new(*otp.MessageSender)),
	),

	wire.NewSet(
		infrawhatsapp.DependencySet,
		wire.Bind(new(otp.WhatsappService), new(*infrawhatsapp.Service)),
	),

	wire.NewSet(
		translation.DependencySet,
		wire.Bind(new(otp.TranslationService), new(*translation.Service)),
		wire.Bind(new(featurepasskey.TranslationService), new(*translation.Service)),
	),

	wire.NewSet(
		web.DependencySet,
		wire.Bind(new(translation.StaticAssetResolver), new(*web.StaticAssetResolver)),
	),

	wire.NewSet(
		ratelimit.DependencySet,
		wire.Bind(new(interaction.RateLimiter), new(*ratelimit.Limiter)),
		wire.Bind(new(workflow.RateLimiter), new(*ratelimit.Limiter)),
		wire.Bind(new(authenticationflow.RateLimiter), new(*ratelimit.Limiter)),
		wire.Bind(new(accountmanagement.RateLimitMiddlewareRateLimiter), new(*ratelimit.Limiter)),
		wire.Bind(new(authenticatorservice.RateLimiter), new(*ratelimit.Limiter)),
		wire.Bind(new(otp.RateLimiter), new(*ratelimit.Limiter)),
		wire.Bind(new(messaging.RateLimiter), new(*ratelimit.Limiter)),
		wire.Bind(new(mfa.RateLimiter), new(*ratelimit.Limiter)),
		wire.Bind(new(featuresiwe.RateLimiter), new(*ratelimit.Limiter)),
	),

	wire.NewSet(
		lockout.DependencySet,
		wire.Bind(new(authenticatorservice.LockoutProvider), new(*lockout.Service)),
		wire.Bind(new(mfa.LockoutProvider), new(*lockout.Service)),
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
		wire.Bind(new(workflow.StdAttrsService), new(*featurestdattrs.Service)),
		wire.Bind(new(authenticationflow.StdAttrsService), new(*featurestdattrs.Service)),
		wire.Bind(new(hook.StandardAttributesServiceNoEvent), new(*featurestdattrs.ServiceNoEvent)),
		wire.Bind(new(userimport.StandardAttributesService), new(*featurestdattrs.ServiceNoEvent)),
	),

	presign.DependencySet,

	tutorial.DependencySet,

	sessionlisting.DependencySet,

	wire.NewSet(
		usage.DependencySet,
		wire.Bind(new(messaging.UsageLimiter), new(*usage.Limiter)),
	),

	wire.NewSet(
		sms.DependencySet,
	),

	wire.NewSet(
		messaging.DependencySet,
		wire.Bind(new(otp.Sender), new(*messaging.Sender)),
	),

	wire.NewSet(
		deprecated_infracaptcha.DependencySet,
	),

	wire.NewSet(
		featurepasskey.DependencySet,
		wire.Bind(new(identitypasskey.PasskeyService), new(*featurepasskey.Service)),
		wire.Bind(new(authenticatorpasskey.PasskeyService), new(*featurepasskey.Service)),
		wire.Bind(new(interaction.PasskeyService), new(*featurepasskey.Service)),
		wire.Bind(new(authenticationflow.PasskeyRequestOptionsService), new(*featurepasskey.RequestOptionsService)),
		wire.Bind(new(authenticationflow.PasskeyCreationOptionsService), new(*featurepasskey.CreationOptionsService)),
		wire.Bind(new(authenticationflow.PasskeyService), new(*featurepasskey.Service)),
	),

	wire.NewSet(
		featuresiwe.DependencySet,
		wire.Bind(new(identitysiwe.SIWEService), new(*featuresiwe.Service)),
	),

	wire.NewSet(
		featureweb3.DependencySet,
		wire.Bind(new(user.Web3Service), new(*featureweb3.Service)),
	),

	wire.NewSet(
		workflow.DependencySet,
	),

	wire.NewSet(
		authenticationflow.DependencySet,
	),

	wire.NewSet(
		accountmanagement.DependencySet,
	),

	wire.NewSet(
		accountmigration.DependencySet,
		wire.Bind(new(workflow.AccountMigrationService), new(*accountmigration.Service)),
		wire.Bind(new(authenticationflow.AccountMigrationService), new(*accountmigration.Service)),
	),

	wire.NewSet(
		app2app.DependencySet,
		wire.Bind(new(oauthhandler.App2AppService), new(*app2app.Provider)),
	),

	wire.NewSet(
		tester.DependencySet,
		wire.Bind(new(webapp.TesterService), new(*tester.TesterStore)),
	),

	wire.NewSet(
		oauthclient.DependencySet,
		wire.Bind(new(oauthhandler.OAuthClientResolver), new(*oauthclient.Resolver)),
		wire.Bind(new(oidc.UIInfoClientResolver), new(*oauthclient.Resolver)),
		wire.Bind(new(webapp.WebappOAuthClientResolver), new(*oauthclient.Resolver)),
		wire.Bind(new(interaction.OAuthClientResolver), new(*oauthclient.Resolver)),
		wire.Bind(new(oauth.OAuthClientResolver), new(*oauthclient.Resolver)),
		wire.Bind(new(authenticationflow.OAuthClientResolver), new(*oauthclient.Resolver)),
	),

	userimport.DependencySet,

	wire.NewSet(
		endpoints.DependencySet,
		wire.Bind(new(oauth.BaseURLProvider), new(*endpoints.Endpoints)),
		wire.Bind(new(oauth.EndpointsProvider), new(*endpoints.Endpoints)),
		wire.Bind(new(oidc.BaseURLProvider), new(*endpoints.Endpoints)),
		wire.Bind(new(oidc.EndpointsProvider), new(*endpoints.Endpoints)),
		wire.Bind(new(oidc.UIURLBuilderAuthUIEndpointsProvider), new(*endpoints.Endpoints)),
		wire.Bind(new(otp.EndpointsProvider), new(*endpoints.Endpoints)),
		wire.Bind(new(tester.EndpointsProvider), new(*endpoints.Endpoints)),
		wire.Bind(new(interaction.OAuthRedirectURIBuilder), new(*endpoints.Endpoints)),
	),

	wire.NewSet(
		redisqueue.ProducerDependencySet,
		wire.Bind(new(libes.UserReindexCreateProducer), new(*redisqueue.UserReindexProducer)),
	),

	wire.NewSet(
		webappoauth.DependencySet,
		wire.Bind(new(interaction.OAuthStateStore), new(*webappoauth.Store)),
	),
)
