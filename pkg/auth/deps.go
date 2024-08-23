package auth

import (
	"net/http"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/auth/api"
	handlerapi "github.com/authgear/authgear-server/pkg/auth/handler/api"
	handleroauth "github.com/authgear/authgear-server/pkg/auth/handler/oauth"
	handlersaml "github.com/authgear/authgear-server/pkg/auth/handler/saml"
	handlersiwe "github.com/authgear/authgear-server/pkg/auth/handler/siwe"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	handlerwebappauthflowv2 "github.com/authgear/authgear-server/pkg/auth/handler/webapp/authflowv2"
	viewmodelswebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/accountmanagement"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/password"
	authenticatorservice "github.com/authgear/authgear-server/pkg/lib/authn/authenticator/service"
	"github.com/authgear/authgear-server/pkg/lib/authn/challenge"
	identityanonymous "github.com/authgear/authgear-server/pkg/lib/authn/identity/anonymous"
	identityservice "github.com/authgear/authgear-server/pkg/lib/authn/identity/service"
	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/endpoints"
	"github.com/authgear/authgear-server/pkg/lib/facade"
	featurecustomattrs "github.com/authgear/authgear-server/pkg/lib/feature/customattrs"
	"github.com/authgear/authgear-server/pkg/lib/feature/forgotpassword"
	featurepasskey "github.com/authgear/authgear-server/pkg/lib/feature/passkey"
	featuresiwe "github.com/authgear/authgear-server/pkg/lib/feature/siwe"
	featurestdattrs "github.com/authgear/authgear-server/pkg/lib/feature/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/infra/middleware"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/meter"
	"github.com/authgear/authgear-server/pkg/lib/nonce"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	oauthhandler "github.com/authgear/authgear-server/pkg/lib/oauth/handler"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oauthsession"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	oidchandler "github.com/authgear/authgear-server/pkg/lib/oauth/oidc/handler"
	"github.com/authgear/authgear-server/pkg/lib/oauthclient"
	"github.com/authgear/authgear-server/pkg/lib/presign"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlsession"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/sessionlisting"
	"github.com/authgear/authgear-server/pkg/lib/tester"
	"github.com/authgear/authgear-server/pkg/lib/translation"
	"github.com/authgear/authgear-server/pkg/lib/web"
	"github.com/authgear/authgear-server/pkg/lib/webappoauth"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/intl"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/template"
)

func ProvideOAuthMetadataProviders(oauth *oauth.MetadataProvider, oidc *oidc.MetadataProvider) []handleroauth.MetadataProvider {
	return []handleroauth.MetadataProvider{oauth, oidc}
}

var DependencySet = wire.NewSet(
	deps.RequestDependencySet,
	deps.CommonDependencySet,

	nonce.DependencySet,
	wire.Bind(new(interaction.NonceService), new(*nonce.Service)),

	wire.Bind(new(webapp.SessionMiddlewareOAuthSessionService), new(*oauthsession.StoreRedis)),
	wire.Bind(new(webapp.SessionMiddlewareOAuthUIInfoResolver), new(*oidc.UIInfoResolver)),
	wire.Bind(new(webapp.GraphService), new(*interaction.Service)),
	wire.Bind(new(webapp.CookieManager), new(*httputil.CookieManager)),
	wire.Bind(new(webapp.OAuthClientResolver), new(*oauthclient.Resolver)),
	wire.Bind(new(webapp.TutorialMiddlewareTutorialCookie), new(*httputil.TutorialCookie)),
	wire.Bind(new(webapp.AuthflowNavigatorOAuthStateStore), new(*webappoauth.Store)),
	wire.Bind(new(handlerwebappauthflowv2.AuthflowV2NavigatorOAuthStateStore), new(*webappoauth.Store)),
	wire.Bind(new(handlerwebapp.CookieManager), new(*httputil.CookieManager)),
	wire.Bind(new(handlerwebapp.AuthflowControllerCookieManager), new(*httputil.CookieManager)),
	wire.Bind(new(handlerwebapp.AuthflowControllerOAuthSessionService), new(*oauthsession.StoreRedis)),
	wire.Bind(new(handlerwebapp.AuthflowControllerUIInfoResolver), new(*oidc.UIInfoResolver)),
	wire.Bind(new(oauthhandler.CookieManager), new(*httputil.CookieManager)),
	wire.Bind(new(oauth.AppSessionTokenServiceCookieManager), new(*httputil.CookieManager)),
	wire.Bind(new(handlerwebapp.TutorialCookie), new(*httputil.TutorialCookie)),
	wire.Bind(new(handlerapi.WorkflowNewCookieManager), new(*httputil.CookieManager)),
	wire.Bind(new(handlerapi.WorkflowInputCookieManager), new(*httputil.CookieManager)),
	wire.Bind(new(handlerapi.WorkflowGetCookieManager), new(*httputil.CookieManager)),
	wire.Bind(new(handlerapi.WorkflowNewOAuthSessionService), new(*oauthsession.StoreRedis)),
	wire.Bind(new(handlerapi.WorkflowNewUIInfoResolver), new(*oidc.UIInfoResolver)),
	wire.Bind(new(handlerapi.WorkflowV2CookieManager), new(*httputil.CookieManager)),
	wire.Bind(new(handlerapi.WorkflowV2OAuthSessionService), new(*oauthsession.StoreRedis)),
	wire.Bind(new(handlerapi.WorkflowV2UIInfoResolver), new(*oidc.UIInfoResolver)),
	wire.Bind(new(handlerapi.AuthenticationFlowV1CookieManager), new(*httputil.CookieManager)),
	wire.Bind(new(handlerapi.AuthenticationFlowV1OAuthSessionService), new(*oauthsession.StoreRedis)),
	wire.Bind(new(handlerapi.AuthenticationFlowV1UIInfoResolver), new(*oidc.UIInfoResolver)),

	wire.Bind(new(webapp.SessionMiddlewareSAMLSessionService), new(*samlsession.StoreRedis)),
	wire.Bind(new(handlerwebapp.AuthflowControllerSAMLSessionService), new(*samlsession.StoreRedis)),
	wire.Bind(new(webapp.SessionMiddlewareSAMLUIInfoResolver), new(*samlsession.UIService)),

	wire.Bind(new(webapp.UIInfoResolver), new(*authenticationinfo.UIService)),
	wire.Bind(new(handlerwebapp.SettingsDeleteAccountSuccessUIInfoResolver), new(*authenticationinfo.UIService)),
	wire.Bind(new(handlerwebapp.SelectAccountAuthenticationInfoService), new(*authenticationinfo.StoreRedis)),
	wire.Bind(new(handlerwebappauthflowv2.SelectAccountAuthenticationInfoService), new(*authenticationinfo.StoreRedis)),
	wire.Bind(new(handlerwebapp.SettingsDeleteAccountSuccessAuthenticationInfoService), new(*authenticationinfo.StoreRedis)),
	wire.Bind(new(handlerwebapp.SettingsDeleteAccountAuthenticationInfoService), new(*authenticationinfo.StoreRedis)),
	wire.Bind(new(handlerwebapp.SetupTOTPEndpointsProvider), new(*endpoints.Endpoints)),
	wire.Bind(new(handlerwebapp.OAuthEntrypointEndpointsProvider), new(*endpoints.Endpoints)),
	wire.Bind(new(handlerwebapp.ConfirmTerminateOtherSessionsEndpointsProvider), new(*endpoints.Endpoints)),
	wire.Bind(new(handlerwebapp.AuthflowLoginEndpointsProvider), new(*endpoints.Endpoints)),
	wire.Bind(new(handlerwebapp.PanicMiddlewareEndpointsProvider), new(*endpoints.Endpoints)),
	wire.Bind(new(webapp.AuthflowNavigatorEndpointsProvider), new(*endpoints.Endpoints)),
	wire.Bind(new(webapp.SuccessPageMiddlewareEndpointsProvider), new(*endpoints.Endpoints)),
	wire.Bind(new(handlerwebappauthflowv2.AuthflowLoginEndpointsProvider), new(*endpoints.Endpoints)),
	wire.Bind(new(handlerwebappauthflowv2.AuthflowV2NavigatorEndpointsProvider), new(*endpoints.Endpoints)),
	wire.Bind(new(handlerwebappauthflowv2.AuthflowV2PromoteEndpointsProvider), new(*endpoints.Endpoints)),
	wire.Bind(new(handlerwebapp.AuthflowSignupEndpointsProvider), new(*endpoints.Endpoints)),
	wire.Bind(new(handlerwebapp.AuthflowPromoteEndpointsProvider), new(*endpoints.Endpoints)),
	wire.Bind(new(oidchandler.WebAppURLsProvider), new(*endpoints.Endpoints)),

	webapp.DependencySet,
	wire.Bind(new(handlerwebapp.AnonymousUserPromotionService), new(*webapp.AnonymousUserPromotionService)),

	wire.Bind(new(webapp.AnonymousIdentityProvider), new(*identityanonymous.Provider)),

	middleware.DependencySet,
	wire.Bind(new(webapp.SettingsSubRoutesMiddlewareIdentityService), new(*facade.IdentityFacade)),

	handleroauth.DependencySet,
	wire.Bind(new(handleroauth.ProtocolAuthorizeHandler), new(*oauthhandler.AuthorizationHandler)),
	wire.Bind(new(handleroauth.ProtocolConsentHandler), new(*oauthhandler.AuthorizationHandler)),
	wire.Bind(new(handleroauth.ProtocolTokenHandler), new(*oauthhandler.TokenHandler)),
	wire.Bind(new(handleroauth.ProtocolRevokeHandler), new(*oauthhandler.RevokeHandler)),
	wire.Bind(new(handleroauth.ProtocolEndSessionHandler), new(*oidchandler.EndSessionHandler)),
	wire.Bind(new(handleroauth.ProtocolUserInfoProvider), new(*oidc.IDTokenIssuer)),
	wire.Bind(new(handleroauth.JWSSource), new(*oidc.IDTokenIssuer)),
	wire.Bind(new(handleroauth.ChallengeProvider), new(*challenge.Provider)),
	wire.Bind(new(handleroauth.JSONResponseWriter), new(*httputil.JSONResponseWriter)),
	wire.Bind(new(handleroauth.AppSessionTokenIssuer), new(*oauthhandler.TokenHandler)),
	wire.Bind(new(handleroauth.Renderer), new(*handlerwebapp.ResponseRenderer)),
	wire.Bind(new(handleroauth.ProtocolIdentityService), new(*identityservice.Service)),
	wire.Bind(new(handleroauth.ProtocolProxyRedirectHandler), new(*oauthhandler.ProxyRedirectHandler)),
	wire.Bind(new(handleroauth.OAuthClientResolver), new(*oauthclient.Resolver)),
	wire.Bind(new(handleroauth.ConsentUserService), new(*user.Queries)),
	ProvideOAuthMetadataProviders,

	handlersaml.DependencySet,

	handlerapi.DependencySet,
	wire.Bind(new(handlerapi.JSONResponseWriter), new(*httputil.JSONResponseWriter)),
	wire.Bind(new(handlerapi.TurboResponseWriter), new(*handlerwebapp.ResponseWriter)),
	wire.Bind(new(handlerapi.AnonymousUserHandler), new(*oauthhandler.AnonymousUserHandler)),
	wire.Bind(new(handlerapi.PromotionCodeIssuer), new(*oauthhandler.AnonymousUserHandler)),
	wire.Bind(new(handlerapi.RateLimiter), new(*ratelimit.Limiter)),
	wire.Bind(new(handlerapi.PresignProvider), new(*presign.Provider)),
	wire.Bind(new(handlerapi.WorkflowNewWorkflowService), new(*workflow.Service)),
	wire.Bind(new(handlerapi.WorkflowGetWorkflowService), new(*workflow.Service)),
	wire.Bind(new(handlerapi.WorkflowInputWorkflowService), new(*workflow.Service)),
	wire.Bind(new(handlerapi.WorkflowV2WorkflowService), new(*workflow.Service)),
	wire.Bind(new(handlerapi.AuthenticationFlowV1WorkflowService), new(*authenticationflow.Service)),
	wire.Bind(new(handlerapi.WorkflowWebsocketEventStore), new(*workflow.EventStoreImpl)),
	wire.Bind(new(handlerapi.WorkflowWebsocketOriginMatcher), new(*middleware.CORSMatcher)),
	wire.Bind(new(handlerapi.AuthenticationFlowV1WebsocketEventStore), new(*authenticationflow.WebsocketEventStore)),
	wire.Bind(new(handlerapi.AuthenticationFlowV1WebsocketOriginMatcher), new(*middleware.CORSMatcher)),
	wire.Bind(new(handlerapi.AccountManagementV1IdentificationHandlerService), new(*accountmanagement.Service)),
	wire.Bind(new(handlerapi.AccountManagementV1IdentificationOAuthHandlerService), new(*accountmanagement.Service)),

	viewmodelswebapp.DependencySet,
	wire.Bind(new(viewmodelswebapp.StaticAssetResolver), new(*web.StaticAssetResolver)),
	wire.Bind(new(viewmodelswebapp.ErrorService), new(*webapp.ErrorService)),
	wire.Bind(new(viewmodelswebapp.TranslationService), new(*translation.Service)),
	wire.Bind(new(viewmodelswebapp.FlashMessage), new(*httputil.FlashMessage)),
	wire.Bind(new(viewmodelswebapp.SettingsIdentityService), new(*identityservice.Service)),
	wire.Bind(new(viewmodelswebapp.SettingsAuthenticatorService), new(*authenticatorservice.Service)),
	wire.Bind(new(viewmodelswebapp.SettingsMFAService), new(*mfa.Service)),
	wire.Bind(new(viewmodelswebapp.SettingsProfileUserService), new(*user.Queries)),
	wire.Bind(new(viewmodelswebapp.SettingsProfileIdentityService), new(*facade.IdentityFacade)),
	wire.Bind(new(viewmodelswebapp.WebappOAuthClientResolver), new(*oauthclient.Resolver)),

	handlerwebapp.DependencySet,
	wire.Bind(new(handlerwebapp.AuthflowControllerOAuthClientResolver), new(*oauthclient.Resolver)),
	wire.Bind(new(handlerwebapp.AuthflowControllerSessionStore), new(*webapp.SessionStoreRedis)),
	wire.Bind(new(handlerwebapp.SettingsDeleteAccountSessionStore), new(*webapp.SessionStoreRedis)),
	wire.Bind(new(handlerwebapp.SettingsAuthenticatorService), new(*authenticatorservice.Service)),
	wire.Bind(new(handlerwebapp.SettingsMFAService), new(*mfa.Service)),
	wire.Bind(new(handlerwebapp.SettingsIdentityService), new(*identityservice.Service)),
	wire.Bind(new(handlerwebapp.SettingsVerificationService), new(*verification.Service)),
	wire.Bind(new(handlerwebapp.SettingsSessionManager), new(*session.Manager)),
	wire.Bind(new(handlerwebapp.SettingsProfileEditUserService), new(*facade.UserFacade)),
	wire.Bind(new(handlerwebapp.SettingsProfileEditStdAttrsService), new(*featurestdattrs.Service)),
	wire.Bind(new(handlerwebapp.SettingsProfileEditCustomAttrsService), new(*featurecustomattrs.Service)),
	wire.Bind(new(handlerwebapp.SettingsDeleteAccountUserService), new(*facade.UserFacade)),
	wire.Bind(new(handlerwebapp.SettingsDeleteAccountOAuthSessionService), new(*oauthsession.StoreRedis)),
	wire.Bind(new(handlerwebapp.SettingsAuthorizationService), new(*oauth.AuthorizationService)),
	wire.Bind(new(handlerwebapp.SettingsSessionListingService), new(*sessionlisting.SessionListingService)),
	wire.Bind(new(handlerwebapp.EnterLoginIDService), new(*identityservice.Service)),
	wire.Bind(new(handlerwebapp.PasswordPolicy), new(*password.Checker)),
	wire.Bind(new(handlerwebapp.ResetPasswordService), new(*forgotpassword.Service)),
	wire.Bind(new(handlerwebapp.LogoutSessionManager), new(*session.Manager)),
	wire.Bind(new(handlerwebapp.PageService), new(*webapp.Service2)),
	wire.Bind(new(handlerwebapp.ResourceManager), new(*resource.Manager)),
	wire.Bind(new(handlerwebapp.GlobalEmbeddedResourceManager), new(*web.GlobalEmbeddedResourceManager)),
	wire.Bind(new(handlerwebapp.JSONResponseWriter), new(*httputil.JSONResponseWriter)),
	wire.Bind(new(handlerwebapp.FlashMessage), new(*httputil.FlashMessage)),
	wire.Bind(new(handlerwebapp.SelectAccountIdentityService), new(*identityservice.Service)),
	wire.Bind(new(handlerwebapp.SelectAccountUserService), new(*user.Queries)),
	wire.Bind(new(handlerwebappauthflowv2.SelectAccountIdentityService), new(*identityservice.Service)),
	wire.Bind(new(handlerwebappauthflowv2.SelectAccountUserService), new(*user.Queries)),
	wire.Bind(new(handlerwebapp.MeterService), new(*meter.Service)),
	wire.Bind(new(handlerwebapp.ErrorService), new(*webapp.ErrorService)),
	wire.Bind(new(handlerwebapp.PasskeyCreationOptionsService), new(*featurepasskey.CreationOptionsService)),
	wire.Bind(new(handlerwebapp.PasskeyRequestOptionsService), new(*featurepasskey.RequestOptionsService)),
	wire.Bind(new(handlerwebapp.WorkflowWebsocketEventStore), new(*workflow.EventStoreImpl)),
	wire.Bind(new(handlerwebapp.AuthenticationFlowWebsocketEventStore), new(*authenticationflow.WebsocketEventStore)),
	wire.Bind(new(handlerwebappauthflowv2.AuthenticationFlowWebsocketEventStore), new(*authenticationflow.WebsocketEventStore)),
	wire.Bind(new(handlerwebapp.TesterAuthTokensIssuer), new(*oauthhandler.TokenHandler)),
	wire.Bind(new(handlerwebapp.TesterCookieManager), new(*httputil.CookieManager)),
	wire.Bind(new(handlerwebapp.TesterAppSessionTokenService), new(*oauth.AppSessionTokenService)),
	wire.Bind(new(handlerwebapp.TesterUserInfoProvider), new(*oidc.IDTokenIssuer)),
	wire.Bind(new(handlerwebapp.TesterOfflineGrantService), new(*oauth.OfflineGrantService)),
	wire.Bind(new(handlerwebapp.AuthflowControllerAuthflowService), new(*authenticationflow.Service)),
	wire.Bind(new(handlerwebapp.AuthflowWechatHandlerOAuthStateStore), new(*webappoauth.Store)),
	wire.Bind(new(handlerwebapp.WechatCallbackHandlerOAuthStateStore), new(*webappoauth.Store)),
	wire.Bind(new(handlerwebapp.SSOCallbackHandlerOAuthStateStore), new(*webappoauth.Store)),
	wire.Bind(new(handlerwebappauthflowv2.AuthflowV2WechatHandlerOAuthStateStore), new(*webappoauth.Store)),

	handlerwebappauthflowv2.DependencySet,

	handlersiwe.DependencySet,
	wire.Bind(new(handlersiwe.NonceHandlerJSONResponseWriter), new(*httputil.JSONResponseWriter)),
	wire.Bind(new(handlersiwe.NonceHandlerSIWEService), new(*featuresiwe.Service)),

	api.DependencySet,
	wire.Bind(new(api.JSONResponseWriter), new(*httputil.JSONResponseWriter)),
	wire.Bind(new(authenticationflow.JSONResponseWriter), new(*httputil.JSONResponseWriter)),
	wire.Bind(new(accountmanagement.RateLimitMiddlewareJSONResponseWriter), new(*httputil.JSONResponseWriter)),
)

func ProvideOAuthConfig() *config.OAuthConfig {
	return &config.OAuthConfig{}
}

func ProvideUIConfig() *config.UIConfig {
	return &config.UIConfig{
		PhoneInput: &config.PhoneInputConfig{},
	}
}

func ProvideUIFeatureConfig() *config.UIFeatureConfig {
	return &config.UIFeatureConfig{
		WhiteLabeling: &config.WhiteLabelingFeatureConfig{},
	}
}

func ProvideForgotPasswordConfig() *config.ForgotPasswordConfig {
	c := &config.ForgotPasswordConfig{}
	config.SetFieldDefaults(c)
	return c
}

func ProvideAuthenticationConfig() *config.AuthenticationConfig {
	c := &config.AuthenticationConfig{}
	c.SetDefaults()
	return c
}

func ProvideGoogleTagManagerConfig() *config.GoogleTagManagerConfig {
	return &config.GoogleTagManagerConfig{}
}

func ProvideBotProtectionConfig() *config.BotProtectionConfig {
	return &config.BotProtectionConfig{}
}

func ProvideLocalizationConfig(defaultLang template.DefaultLanguageTag, supported template.SupportedLanguageTags) *config.LocalizationConfig {
	defaultLangStr := string(defaultLang)
	return &config.LocalizationConfig{
		FallbackLanguage:   &defaultLangStr,
		SupportedLanguages: []string(supported),
	}
}

func ProvideCookieManager(r *http.Request, trustProxy config.TrustProxy) *httputil.CookieManager {
	m := &httputil.CookieManager{
		Request:    r,
		TrustProxy: bool(trustProxy),
	}
	return m
}

type NoopErrorService struct{}

func (*NoopErrorService) PopError(w http.ResponseWriter, r *http.Request) (*webapp.ErrorState, bool) {
	return nil, false
}

func ProvideNoopErrorService() *NoopErrorService {
	return &NoopErrorService{}
}

var RequestMiddlewareDependencySet = wire.NewSet(
	template.DependencySet,
	web.DependencySet,
	translation.DependencySet,
	deps.RootDependencySet,
	httputil.DependencySet,

	ProvideOAuthConfig,
	ProvideUIConfig,
	ProvideUIFeatureConfig,
	ProvideForgotPasswordConfig,
	ProvideAuthenticationConfig,
	ProvideGoogleTagManagerConfig,
	ProvideLocalizationConfig,
	ProvideBotProtectionConfig,

	ProvideCookieManager,

	deps.ProvideRequestContext,
	deps.ProvideRemoteIP,
	deps.ProvideUserAgentString,
	deps.ProvideHTTPHost,
	deps.ProvideHTTPProto,

	wire.Value(template.DefaultLanguageTag(intl.BuiltinBaseLanguage)),
	wire.Value(template.SupportedLanguageTags([]string{intl.BuiltinBaseLanguage})),

	viewmodelswebapp.NewBaseLogger,
	wire.Struct(new(viewmodelswebapp.BaseViewModeler), "*"),
	wire.Struct(new(deps.RequestMiddleware), "*"),

	wire.Bind(new(template.ResourceManager), new(*resource.Manager)),
	wire.Bind(new(web.ResourceManager), new(*resource.Manager)),

	wire.Bind(new(viewmodelswebapp.StaticAssetResolver), new(*web.StaticAssetResolver)),
	wire.Bind(new(translation.StaticAssetResolver), new(*web.StaticAssetResolver)),
	wire.Bind(new(web.EmbeddedResourceManager), new(*web.GlobalEmbeddedResourceManager)),

	wire.Bind(new(viewmodelswebapp.TranslationService), new(*translation.Service)),

	ProvideNoopErrorService,
	wire.Bind(new(viewmodelswebapp.ErrorService), new(*NoopErrorService)),

	wire.Bind(new(webapp.CookieManager), new(*httputil.CookieManager)),
	wire.Bind(new(viewmodelswebapp.FlashMessage), new(*httputil.FlashMessage)),
	wire.Bind(new(httputil.FlashMessageCookieManager), new(*httputil.CookieManager)),

	endpoints.DependencySet,
	wire.Bind(new(tester.EndpointsProvider), new(*endpoints.Endpoints)),

	oauthclient.DependencySet,
	wire.Bind(new(viewmodelswebapp.WebappOAuthClientResolver), new(*oauthclient.Resolver)),
)

func RequestMiddleware(p *deps.RootProvider, configSource *configsource.ConfigSource, factory func(http.ResponseWriter, *http.Request, *deps.RootProvider, *configsource.ConfigSource) httproute.Middleware) httproute.Middleware {
	return httproute.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			m := factory(w, r, p, configSource)
			h := m.Handle(next)
			h.ServeHTTP(w, r)
		})
	})
}

var AuthflowUIHandlerDependencySet = wire.NewSet(
	wire.Struct(new(webapp.AuthflowNavigator), "*"),
	wire.Bind(new(handlerwebapp.AuthflowNavigator), new(*webapp.AuthflowNavigator)),
)

var AuthflowV2UIHandlerDependencySet = wire.NewSet(
	wire.Struct(new(handlerwebappauthflowv2.AuthflowV2Navigator), "*"),
	wire.Bind(new(handlerwebapp.AuthflowNavigator), new(*handlerwebappauthflowv2.AuthflowV2Navigator)),
	wire.Bind(new(handlerwebappauthflowv2.AuthflowV2ChangePasswordNavigator), new(*handlerwebappauthflowv2.AuthflowV2Navigator)),
)
