package auth

import (
	"net/http"

	oauthhandler "github.com/authgear/authgear-server/pkg/auth/handler/oauth"
	webapphandler "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func NewRouter(p *deps.RootProvider, configSource *configsource.ConfigSource, staticAsset StaticAssetConfig) *httproute.Router {
	router := httproute.NewRouter()

	router.Add(httproute.Route{
		Methods:     []string{"GET"},
		PathPattern: "/healthz",
	}, http.HandlerFunc(httputil.HealthCheckHandler))

	rootChain := httproute.Chain(
		p.RootMiddleware(newPanicEndMiddleware),
		p.RootMiddleware(newBodyLimitMiddleware),
		p.RootMiddleware(newSentryMiddleware),
		&deps.RequestMiddleware{
			RootProvider: p,
			ConfigSource: configSource,
		},
		p.Middleware(newPanicLogMiddleware),
		p.Middleware(newSessionMiddleware),
		p.Middleware(newCORSMiddleware),
	)

	apiChain := httproute.Chain(
		rootChain,
		p.Middleware(newPanicAPIMiddleware),
	)

	scopedChain := httproute.Chain(
		rootChain,
		p.Middleware(newPanicWriteEmptyResponseMiddleware),
		// Current we only require valid session and do not require any scope.
		httproute.MiddlewareFunc(oauth.RequireScope()),
	)

	webappChain := httproute.Chain(
		rootChain,
		httproute.MiddlewareFunc(httputil.NoCache),
		httproute.MiddlewareFunc(webapp.IntlMiddleware),
		p.Middleware(newCSPMiddleware),
		p.Middleware(newPanicWebAppMiddleware),
		p.Middleware(newWebAppSessionMiddleware),
	)
	webappSSOCallbackChain := httproute.Chain(
		webappChain,
	)
	webappPageChain := httproute.Chain(
		webappChain,
		p.Middleware(newCSRFMiddleware),
	)
	webappAuthEntrypointChain := httproute.Chain(
		webappPageChain,
		p.Middleware(newAuthEntryPointMiddleware),
	)
	webappAuthenticatedChain := httproute.Chain(
		webappPageChain,
		webapp.RequireAuthenticatedMiddleware{},
	)

	rootRoute := httproute.Route{Middleware: rootChain}
	apiRoute := httproute.Route{Middleware: apiChain}
	scopedRoute := httproute.Route{Middleware: scopedChain}
	webappRoute := httproute.Route{Middleware: webappChain}
	webappPageRoute := httproute.Route{Middleware: webappPageChain}
	webappAuthEntrypointRoute := httproute.Route{Middleware: webappAuthEntrypointChain}
	webappAuthenticatedRoute := httproute.Route{Middleware: webappAuthenticatedChain}
	webappSSOCallbackRoute := httproute.Route{Middleware: webappSSOCallbackChain}

	router.Add(webapphandler.ConfigureRootRoute(webappAuthEntrypointRoute), p.Handler(newWebAppRootHandler))
	router.Add(webapphandler.ConfigureLoginRoute(webappAuthEntrypointRoute), p.Handler(newWebAppLoginHandler))
	router.Add(webapphandler.ConfigureSignupRoute(webappAuthEntrypointRoute), p.Handler(newWebAppSignupHandler))

	router.Add(webapphandler.ConfigurePromoteRoute(webappPageRoute), p.Handler(newWebAppPromoteHandler))
	router.Add(webapphandler.ConfigureEnterPasswordRoute(webappPageRoute), p.Handler(newWebAppEnterPasswordHandler))
	router.Add(webapphandler.ConfigureSetupTOTPRoute(webappPageRoute), p.Handler(newWebAppSetupTOTPHandler))
	router.Add(webapphandler.ConfigureEnterTOTPRoute(webappPageRoute), p.Handler(newWebAppEnterTOTPHandler))
	router.Add(webapphandler.ConfigureSetupOOBOTPRoute(webappPageRoute), p.Handler(newWebAppSetupOOBOTPHandler))
	router.Add(webapphandler.ConfigureEnterOOBOTPRoute(webappPageRoute), p.Handler(newWebAppEnterOOBOTPHandler))
	router.Add(webapphandler.ConfigureEnterRecoveryCodeRoute(webappPageRoute), p.Handler(newWebAppEnterRecoveryCodeHandler))
	router.Add(webapphandler.ConfigureSetupRecoveryCodeRoute(webappPageRoute), p.Handler(newWebAppSetupRecoveryCodeHandler))
	router.Add(webapphandler.ConfigureVerifyIdentityRoute(webappPageRoute), p.Handler(newWebAppVerifyIdentityHandler))
	router.Add(webapphandler.ConfigureVerifyIdentitySuccessRoute(webappPageRoute), p.Handler(newWebAppVerifyIdentitySuccessHandler))
	router.Add(webapphandler.ConfigureCreatePasswordRoute(webappPageRoute), p.Handler(newWebAppCreatePasswordHandler))
	router.Add(webapphandler.ConfigureForgotPasswordRoute(webappPageRoute), p.Handler(newWebAppForgotPasswordHandler))
	router.Add(webapphandler.ConfigureForgotPasswordSuccessRoute(webappPageRoute), p.Handler(newWebAppForgotPasswordSuccessHandler))
	router.Add(webapphandler.ConfigureResetPasswordRoute(webappPageRoute), p.Handler(newWebAppResetPasswordHandler))
	router.Add(webapphandler.ConfigureResetPasswordSuccessRoute(webappPageRoute), p.Handler(newWebAppResetPasswordSuccessHandler))

	router.Add(webapphandler.ConfigureLogoutRoute(webappAuthenticatedRoute), p.Handler(newWebAppLogoutHandler))
	router.Add(webapphandler.ConfigureEnterLoginIDRoute(webappAuthenticatedRoute), p.Handler(newWebAppEnterLoginIDHandler))
	router.Add(webapphandler.ConfigureSettingsIdentityRoute(webappAuthenticatedRoute), p.Handler(newWebAppSettingsIdentityHandler))
	router.Add(webapphandler.ConfigureSettingsTOTPRoute(webappAuthenticatedRoute), p.Handler(newWebAppSettingsTOTPHandler))
	router.Add(webapphandler.ConfigureSettingsOOBOTPRoute(webappAuthenticatedRoute), p.Handler(newWebAppSettingsOOBOTPHandler))
	router.Add(webapphandler.ConfigureSettingsRecoveryCodeRoute(webappAuthenticatedRoute), p.Handler(newWebAppSettingsRecoveryCodeHandler))
	router.Add(webapphandler.ConfigureSettingsRoute(webappAuthenticatedRoute), p.Handler(newWebAppSettingsHandler))
	router.Add(webapphandler.ConfigureChangePasswordRoute(webappAuthenticatedRoute), p.Handler(newWebAppChangePasswordHandler))
	router.Add(webapphandler.ConfigureChangeSecondaryPasswordRoute(webappAuthenticatedRoute), p.Handler(newWebAppChangeSecondaryPasswordHandler))

	router.Add(webapphandler.ConfigureSSOCallbackRoute(webappSSOCallbackRoute), p.Handler(newWebAppSSOCallbackHandler))

	router.Add(oauthhandler.ConfigureOIDCMetadataRoute(rootRoute), p.Handler(newOAuthMetadataHandler))
	router.Add(oauthhandler.ConfigureOAuthMetadataRoute(rootRoute), p.Handler(newOAuthMetadataHandler))
	router.Add(oauthhandler.ConfigureJWKSRoute(rootRoute), p.Handler(newOAuthJWKSHandler))
	router.Add(oauthhandler.ConfigureAuthorizeRoute(rootRoute), p.Handler(newOAuthAuthorizeHandler))
	router.Add(oauthhandler.ConfigureTokenRoute(rootRoute), p.Handler(newOAuthTokenHandler))
	router.Add(oauthhandler.ConfigureRevokeRoute(rootRoute), p.Handler(newOAuthRevokeHandler))
	router.Add(oauthhandler.ConfigureEndSessionRoute(rootRoute), p.Handler(newOAuthEndSessionHandler))
	router.Add(oauthhandler.ConfigureChallengeRoute(apiRoute), p.Handler(newOAuthChallengeHandler))

	router.Add(oauthhandler.ConfigureUserInfoRoute(scopedRoute), p.Handler(newOAuthUserInfoHandler))

	if staticAsset.ServingEnabled {
		router.Add(webapphandler.ConfigureStaticAssetsRoute(webappRoute), p.Handler(newWebAppStaticAssetsHandler))
	}

	return router
}
