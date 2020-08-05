package server

import (
	"net/http"

	configsource "github.com/authgear/authgear-server/pkg/auth/config/source"
	"github.com/authgear/authgear-server/pkg/auth/dependency/oauth"
	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/internalserver"
	oauthhandler "github.com/authgear/authgear-server/pkg/auth/handler/oauth"
	webapphandler "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/deps"
	"github.com/authgear/authgear-server/pkg/httproute"
	"github.com/authgear/authgear-server/pkg/httputil"
)

func setupInternalRoutes(p *deps.RootProvider, configSource configsource.Source) *httproute.Router {
	router := httproute.NewRouter()

	router.Add(httproute.Route{
		Methods:     []string{"GET"},
		PathPattern: "/healthz",
	}, http.HandlerFunc(httputil.HealthCheckHandler))

	chain := httproute.Chain(
		p.RootMiddleware(newRootRecoverMiddleware),
		p.RootMiddleware(newSentryMiddleware),
		&deps.RequestMiddleware{
			RootProvider: p,
			ConfigSource: configSource,
		},
		p.Middleware(newRequestRecoverMiddleware),
		p.Middleware(newSessionMiddleware),
	)

	route := httproute.Route{Middleware: chain}

	router.Add(internalserver.ConfigureResolveRoute(route), p.Handler(newSessionResolveHandler))

	return router
}

func setupRoutes(p *deps.RootProvider, configSource configsource.Source) *httproute.Router {
	router := httproute.NewRouter()

	router.Add(httproute.Route{
		Methods:     []string{"GET"},
		PathPattern: "/healthz",
	}, http.HandlerFunc(httputil.HealthCheckHandler))

	rootChain := httproute.Chain(
		p.RootMiddleware(newRootRecoverMiddleware),
		p.RootMiddleware(newSentryMiddleware),
		&deps.RequestMiddleware{
			RootProvider: p,
			ConfigSource: configSource,
		},
		p.Middleware(newRequestRecoverMiddleware),
		p.Middleware(newSessionMiddleware),
		p.Middleware(newCORSMiddleware),
	)
	webappSSOCallbackChain := httproute.Chain(
		rootChain,
		httproute.MiddlewareFunc(webapp.PostNoCacheMiddleware),
	)
	scopedChain := httproute.Chain(
		rootChain,
		// Current we only require valid session and do not require any scope.
		httproute.MiddlewareFunc(oauth.RequireScope()),
	)

	webappChain := httproute.Chain(
		rootChain,
		httproute.MiddlewareFunc(webapp.IntlMiddleware),
		p.Middleware(newCSPMiddleware),
		p.Middleware(newCSRFMiddleware),
		httproute.MiddlewareFunc(webapp.PostNoCacheMiddleware),
		p.Middleware(newWebAppStateMiddleware),
	)
	webappAuthEntrypointChain := httproute.Chain(
		webappChain,
		p.Middleware(newAuthEntryPointMiddleware),
	)
	webappAuthenticatedChain := httproute.Chain(
		webappChain,
		webapp.RequireAuthenticatedMiddleware{},
	)

	rootRoute := httproute.Route{Middleware: rootChain}
	scopedRoute := httproute.Route{Middleware: scopedChain}
	webappRoute := httproute.Route{Middleware: webappChain}
	webappAuthEntrypointRoute := httproute.Route{Middleware: webappAuthEntrypointChain}
	webappAuthenticatedRoute := httproute.Route{Middleware: webappAuthenticatedChain}
	webappSSOCallbackRoute := httproute.Route{Middleware: webappSSOCallbackChain}

	router.Add(webapphandler.ConfigureRootRoute(webappAuthEntrypointRoute), p.Handler(newWebAppRootHandler))
	router.Add(webapphandler.ConfigureLoginRoute(webappAuthEntrypointRoute), p.Handler(newWebAppLoginHandler))
	router.Add(webapphandler.ConfigureSignupRoute(webappAuthEntrypointRoute), p.Handler(newWebAppSignupHandler))

	router.Add(webapphandler.ConfigurePromoteRoute(webappRoute), p.Handler(newWebAppPromoteHandler))
	router.Add(webapphandler.ConfigureEnterPasswordRoute(webappRoute), p.Handler(newWebAppEnterPasswordHandler))
	router.Add(webapphandler.ConfigureEnterLoginIDRoute(webappRoute), p.Handler(newWebAppEnterLoginIDHandler))
	router.Add(webapphandler.ConfigureSetupOOBOTPRoute(webappRoute), p.Handler(newWebAppSetupOOBOTPHandler))
	router.Add(webapphandler.ConfigureEnterOOBOTPRoute(webappRoute), p.Handler(newWebAppEnterOOBOTPHandler))
	router.Add(webapphandler.ConfigureVerifyUserRoute(webappRoute), p.Handler(newWebAppVerifyUserHandler))
	router.Add(webapphandler.ConfigureVerifyUserSuccessRoute(webappRoute), p.Handler(newWebAppVerifyUserSuccessHandler))
	router.Add(webapphandler.ConfigureCreatePasswordRoute(webappRoute), p.Handler(newWebAppCreatePasswordHandler))
	router.Add(webapphandler.ConfigureForgotPasswordRoute(webappRoute), p.Handler(newWebAppForgotPasswordHandler))
	router.Add(webapphandler.ConfigureForgotPasswordSuccessRoute(webappRoute), p.Handler(newWebAppForgotPasswordSuccessHandler))
	router.Add(webapphandler.ConfigureResetPasswordRoute(webappRoute), p.Handler(newWebAppResetPasswordHandler))
	router.Add(webapphandler.ConfigureResetPasswordSuccessRoute(webappRoute), p.Handler(newWebAppResetPasswordSuccessHandler))

	router.Add(webapphandler.ConfigureLogoutRoute(webappAuthenticatedRoute), p.Handler(newWebAppLogoutHandler))
	router.Add(webapphandler.ConfigureSettingsIdentityRoute(webappAuthenticatedRoute), p.Handler(newWebAppSettingsIdentityHandler))
	router.Add(webapphandler.ConfigureSettingsRoute(webappAuthenticatedRoute), p.Handler(newWebAppSettingsHandler))

	router.Add(webapphandler.ConfigureSSOCallbackRoute(webappSSOCallbackRoute), p.Handler(newWebAppSSOCallbackHandler))

	router.Add(oauthhandler.ConfigureOIDCMetadataRoute(rootRoute), p.Handler(newOAuthMetadataHandler))
	router.Add(oauthhandler.ConfigureOAuthMetadataRoute(rootRoute), p.Handler(newOAuthMetadataHandler))
	router.Add(oauthhandler.ConfigureJWKSRoute(rootRoute), p.Handler(newOAuthJWKSHandler))
	router.Add(oauthhandler.ConfigureAuthorizeRoute(rootRoute), p.Handler(newOAuthAuthorizeHandler))
	router.Add(oauthhandler.ConfigureTokenRoute(rootRoute), p.Handler(newOAuthTokenHandler))
	router.Add(oauthhandler.ConfigureRevokeRoute(rootRoute), p.Handler(newOAuthRevokeHandler))
	router.Add(oauthhandler.ConfigureEndSessionRoute(rootRoute), p.Handler(newOAuthEndSessionHandler))
	router.Add(oauthhandler.ConfigureChallengeRoute(rootRoute), p.Handler(newOAuthChallengeHandler))

	router.Add(oauthhandler.ConfigureUserInfoRoute(scopedRoute), p.Handler(newOAuthUserInfoHandler))

	if p.ServerConfig.StaticAsset.ServingEnabled {
		fileServer := http.FileServer(http.Dir(p.ServerConfig.StaticAsset.Dir))
		staticRoute := httproute.Route{
			Methods:     []string{"HEAD", "GET"},
			PathPattern: "/static/*all",
		}
		router.Add(staticRoute, http.StripPrefix("/static/", fileServer))
	}

	return router
}
