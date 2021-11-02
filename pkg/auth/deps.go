package auth

import (
	"github.com/google/wire"

	handleroauth "github.com/authgear/authgear-server/pkg/auth/handler/oauth"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	viewmodelswebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/analytic"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/password"
	authenticatorservice "github.com/authgear/authgear-server/pkg/lib/authn/authenticator/service"
	"github.com/authgear/authgear-server/pkg/lib/authn/challenge"
	identityanonymous "github.com/authgear/authgear-server/pkg/lib/authn/identity/anonymous"
	identityservice "github.com/authgear/authgear-server/pkg/lib/authn/identity/service"
	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/authn/sso"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/facade"
	"github.com/authgear/authgear-server/pkg/lib/feature/forgotpassword"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/infra/middleware"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/nonce"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	oauthhandler "github.com/authgear/authgear-server/pkg/lib/oauth/handler"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	oidchandler "github.com/authgear/authgear-server/pkg/lib/oauth/oidc/handler"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/translation"
	"github.com/authgear/authgear-server/pkg/lib/web"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

func ProvideOAuthMetadataProviders(oauth *oauth.MetadataProvider, oidc *oidc.MetadataProvider) []handleroauth.MetadataProvider {
	return []handleroauth.MetadataProvider{oauth, oidc}
}

var DependencySet = wire.NewSet(
	deps.RequestDependencySet,
	deps.CommonDependencySet,

	nonce.DependencySet,
	wire.Bind(new(interaction.NonceService), new(*nonce.Service)),

	wire.Bind(new(webapp.GraphService), new(*interaction.Service)),
	wire.Bind(new(webapp.CookieManager), new(*httputil.CookieManager)),
	wire.Bind(new(handlerwebapp.CookieManager), new(*httputil.CookieManager)),
	wire.Bind(new(oauthhandler.CookieManager), new(*httputil.CookieManager)),

	wire.Bind(new(handlerwebapp.SelectAccountAuthenticationInfoService), new(*authenticationinfo.StoreRedis)),

	wire.NewSet(
		wire.Struct(new(MainOriginProvider), "*"),
		wire.Bind(new(OriginProvider), new(*MainOriginProvider)),
		wire.Struct(new(EndpointsProvider), "*"),

		wire.Bind(new(oauth.EndpointsProvider), new(*EndpointsProvider)),
		wire.Bind(new(oauth.BaseURLProvider), new(*EndpointsProvider)),
		wire.Bind(new(webapp.EndpointsProvider), new(*EndpointsProvider)),
		wire.Bind(new(handlerwebapp.SetupTOTPEndpointsProvider), new(*EndpointsProvider)),
		wire.Bind(new(oidc.EndpointsProvider), new(*EndpointsProvider)),
		wire.Bind(new(oidc.BaseURLProvider), new(*EndpointsProvider)),
		wire.Bind(new(sso.EndpointsProvider), new(*EndpointsProvider)),
		wire.Bind(new(otp.EndpointsProvider), new(*EndpointsProvider)),
	),

	webapp.DependencySet,
	wire.Bind(new(oauthhandler.WebAppAuthenticateURLProvider), new(*webapp.AuthenticateURLProvider)),
	wire.Bind(new(oauthhandler.LoginHintHandler), new(*webapp.LoginHintHandler)),
	wire.Bind(new(oidchandler.WebAppURLsProvider), new(*webapp.URLProvider)),
	wire.Bind(new(sso.RedirectURLProvider), new(*webapp.URLProvider)),
	wire.Bind(new(forgotpassword.URLProvider), new(*webapp.URLProvider)),
	wire.Bind(new(verification.WebAppURLProvider), new(*webapp.URLProvider)),
	wire.Bind(new(sso.WechatURLProvider), new(*webapp.WechatURLProvider)),

	wire.Bind(new(webapp.AnonymousIdentityProvider), new(*identityanonymous.Provider)),

	middleware.DependencySet,

	handleroauth.DependencySet,
	wire.Bind(new(handleroauth.ProtocolAuthorizeHandler), new(*oauthhandler.AuthorizationHandler)),
	wire.Bind(new(handleroauth.ProtocolFromWebAppHandler), new(*oauthhandler.AuthorizationHandler)),
	wire.Bind(new(handleroauth.ProtocolTokenHandler), new(*oauthhandler.TokenHandler)),
	wire.Bind(new(handleroauth.ProtocolRevokeHandler), new(*oauthhandler.RevokeHandler)),
	wire.Bind(new(handleroauth.ProtocolEndSessionHandler), new(*oidchandler.EndSessionHandler)),
	wire.Bind(new(handleroauth.ProtocolUserInfoProvider), new(*oidc.IDTokenIssuer)),
	wire.Bind(new(handleroauth.JWSSource), new(*oidc.IDTokenIssuer)),
	wire.Bind(new(handleroauth.ChallengeProvider), new(*challenge.Provider)),
	wire.Bind(new(handleroauth.JSONResponseWriter), new(*httputil.JSONResponseWriter)),
	wire.Bind(new(handleroauth.AppSessionTokenIssuer), new(*oauthhandler.TokenHandler)),
	ProvideOAuthMetadataProviders,

	viewmodelswebapp.DependencySet,
	wire.Bind(new(viewmodelswebapp.StaticAssetResolver), new(*web.StaticAssetResolver)),
	wire.Bind(new(viewmodelswebapp.ErrorCookie), new(*webapp.ErrorCookie)),
	wire.Bind(new(viewmodelswebapp.TranslationService), new(*translation.Service)),
	wire.Bind(new(viewmodelswebapp.FlashMessage), new(*httputil.FlashMessage)),
	wire.Bind(new(viewmodelswebapp.SettingsIdentityService), new(*identityservice.Service)),
	wire.Bind(new(viewmodelswebapp.SettingsAuthenticatorService), new(*authenticatorservice.Service)),
	wire.Bind(new(viewmodelswebapp.SettingsMFAService), new(*mfa.Service)),
	wire.Bind(new(viewmodelswebapp.SettingsUserService), new(*user.Queries)),
	wire.Bind(new(viewmodelswebapp.SettingsProfileUserService), new(*user.Queries)),
	wire.Bind(new(viewmodelswebapp.SettingsProfileIdentityService), new(*facade.IdentityFacade)),

	handlerwebapp.DependencySet,
	wire.Bind(new(handlerwebapp.SettingsAuthenticatorService), new(*authenticatorservice.Service)),
	wire.Bind(new(handlerwebapp.SettingsMFAService), new(*mfa.Service)),
	wire.Bind(new(handlerwebapp.SettingsIdentityService), new(*identityservice.Service)),
	wire.Bind(new(handlerwebapp.SettingsVerificationService), new(*verification.Service)),
	wire.Bind(new(handlerwebapp.SettingsSessionManager), new(*session.Manager)),
	wire.Bind(new(handlerwebapp.SettingsProfileEditUserService), new(*facade.UserFacade)),
	wire.Bind(new(handlerwebapp.EnterLoginIDService), new(*identityservice.Service)),
	wire.Bind(new(handlerwebapp.PasswordPolicy), new(*password.Checker)),
	wire.Bind(new(handlerwebapp.LogoutSessionManager), new(*session.Manager)),
	wire.Bind(new(handlerwebapp.PageService), new(*webapp.Service2)),
	wire.Bind(new(handlerwebapp.ResourceManager), new(*resource.Manager)),
	wire.Bind(new(handlerwebapp.VerifyIdentityVerificationService), new(*verification.Service)),
	wire.Bind(new(handlerwebapp.RateLimiter), new(*ratelimit.Limiter)),
	wire.Bind(new(handlerwebapp.JSONResponseWriter), new(*httputil.JSONResponseWriter)),
	wire.Bind(new(handlerwebapp.FlashMessage), new(*httputil.FlashMessage)),
	wire.Bind(new(handlerwebapp.SelectAccountIdentityService), new(*identityservice.Service)),
	wire.Bind(new(handlerwebapp.SelectAccountUserService), new(*user.Queries)),
	wire.Bind(new(handlerwebapp.AnalyticService), new(*analytic.Service)),
)
