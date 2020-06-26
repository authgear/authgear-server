package main

import (
	"net/http"

	configsource "github.com/skygeario/skygear-server/pkg/auth/config/source"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/webapp"
	"github.com/skygeario/skygear-server/pkg/auth/handler/internalserver"
	oauthhandler "github.com/skygeario/skygear-server/pkg/auth/handler/oauth"
	webapphandler "github.com/skygeario/skygear-server/pkg/auth/handler/webapp"
	"github.com/skygeario/skygear-server/pkg/core/sentry"
	"github.com/skygeario/skygear-server/pkg/deps"
	"github.com/skygeario/skygear-server/pkg/httproute"
	"github.com/skygeario/skygear-server/pkg/httputil"
)

func setupInternalRoutes(p *deps.RootProvider, configSource configsource.Source) *httproute.Router {
	router := httproute.NewRouter()

	router.Add(httproute.Route{
		Methods:     []string{"GET"},
		PathPattern: "/healthz",
	}, http.HandlerFunc(httputil.HealthCheckHandler))

	chain := httproute.Chain(
		p.RootMiddleware(newSentryMiddlewareFactory(sentry.DefaultClient.Hub)),
		p.RootMiddleware(newRecoverMiddleware),
		&deps.RequestMiddleware{
			RootProvider: p,
			ConfigSource: configSource,
		},
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
		p.RootMiddleware(newSentryMiddlewareFactory(sentry.DefaultClient.Hub)),
		p.RootMiddleware(newRecoverMiddleware),
		&deps.RequestMiddleware{
			RootProvider: p,
			ConfigSource: configSource,
		},
		p.Middleware(newSessionMiddleware),
	)
	oauthChain := httproute.Chain(
		rootChain,
		p.Middleware(newCORSMiddleware),
	)
	webappSSOCallbackChain := httproute.Chain(
		rootChain,
		httproute.MiddlewareFunc(webapp.PostNoCacheMiddleware),
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

	oauthRoute := httproute.Route{Middleware: oauthChain}
	webappRoute := httproute.Route{Middleware: webappChain}
	webappAuthEntrypointRoute := httproute.Route{Middleware: webappAuthEntrypointChain}
	webappAuthenticatedRoute := httproute.Route{Middleware: webappAuthenticatedChain}
	webappSSOCallbackRoute := httproute.Route{Middleware: webappSSOCallbackChain}

	router.Add(webapphandler.ConfigureRootRoute(webappAuthEntrypointRoute), p.Handler(newWebAppRootHandler))
	router.Add(webapphandler.ConfigureLoginRoute(webappAuthEntrypointRoute), p.Handler(newWebAppLoginHandler))
	router.Add(webapphandler.ConfigureSignupRoute(webappAuthEntrypointRoute), p.Handler(newWebAppSignupHandler))
	router.Add(webapphandler.ConfigurePromoteRoute(webappAuthEntrypointRoute), p.Handler(newWebAppPromoteHandler))

	router.Add(webapphandler.ConfigureEnterPasswordRoute(webappRoute), p.Handler(newWebAppEnterPasswordHandler))
	router.Add(webapphandler.ConfigureEnterLoginIDRoute(webappRoute), p.Handler(newWebAppEnterLoginIDHandler))
	router.Add(webapphandler.ConfigureOOBOTPRoute(webappRoute), p.Handler(newWebAppOOBOTPHandler))
	router.Add(webapphandler.ConfigureCreatePasswordRoute(webappRoute), p.Handler(newWebAppCreatePasswordHandler))
	router.Add(webapphandler.ConfigureForgotPasswordRoute(webappRoute), p.Handler(newWebAppForgotPasswordHandler))
	router.Add(webapphandler.ConfigureForgotPasswordSuccessRoute(webappRoute), p.Handler(newWebAppForgotPasswordSuccessHandler))
	router.Add(webapphandler.ConfigureResetPasswordRoute(webappRoute), p.Handler(newWebAppResetPasswordHandler))
	router.Add(webapphandler.ConfigureResetPasswordSuccessRoute(webappRoute), p.Handler(newWebAppResetPasswordSuccessHandler))

	router.Add(webapphandler.ConfigureLogoutRoute(webappAuthenticatedRoute), p.Handler(newWebAppLogoutHandler))
	router.Add(webapphandler.ConfigureSettingsIdentityRoute(webappAuthenticatedRoute), p.Handler(newWebAppSettingsIdentityHandler))
	router.Add(webapphandler.ConfigureSettingsRoute(webappAuthenticatedRoute), p.Handler(newWebAppSettingsHandler))

	router.Add(webapphandler.ConfigureSSOCallbackRoute(webappSSOCallbackRoute), p.Handler(newWebAppSSOCallbackHandler))

	router.Add(oauthhandler.ConfigureOIDCMetadataRoute(oauthRoute), p.Handler(newOAuthMetadataHandler))
	router.Add(oauthhandler.ConfigureOAuthMetadataRoute(oauthRoute), p.Handler(newOAuthMetadataHandler))
	router.Add(oauthhandler.ConfigureJWKSRoute(oauthRoute), p.Handler(newOAuthJWKSHandler))
	router.Add(oauthhandler.ConfigureAuthorizeRoute(oauthRoute), p.Handler(newOAuthAuthorizeHandler))
	router.Add(oauthhandler.ConfigureTokenRoute(oauthRoute), p.Handler(newOAuthTokenHandler))
	router.Add(oauthhandler.ConfigureRevokeRoute(oauthRoute), p.Handler(newOAuthRevokeHandler))

	userInfoRoute, userInfoHandler := oauthhandler.ConfigureUserInfoHandler(oauthRoute, p.Handler(newOAuthUserInfoHandler))
	router.Add(userInfoRoute, userInfoHandler)

	router.Add(oauthhandler.ConfigureEndSessionRoute(oauthRoute), p.Handler(newOAuthEndSessionHandler))
	router.Add(oauthhandler.ConfigureChallengeRoute(oauthRoute), p.Handler(newOAuthChallengeHandler))

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
