package main

import (
	"net/http"

	"github.com/gorilla/mux"

	configsource "github.com/skygeario/skygear-server/pkg/auth/config/source"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/webapp"
	oauthhandler "github.com/skygeario/skygear-server/pkg/auth/handler/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/handler/session"
	webapphandler "github.com/skygeario/skygear-server/pkg/auth/handler/webapp"
	"github.com/skygeario/skygear-server/pkg/core/sentry"
	"github.com/skygeario/skygear-server/pkg/deps"
	"github.com/skygeario/skygear-server/pkg/httputil"
)

func NewRouter(p *deps.RootProvider) *mux.Router {
	rootRouter := mux.NewRouter()
	rootRouter.Use(p.RootMiddleware(newSentryMiddlewareFactory(sentry.DefaultClient.Hub)))
	rootRouter.Use(p.RootMiddleware(newRecoverMiddleware))
	return rootRouter
}

func setupRoutes(p *deps.RootProvider, configSource configsource.Source) *mux.Router {
	var router *mux.Router
	var rootRouter *mux.Router
	var webappRouter *mux.Router
	var oauthRouter *mux.Router

	router = NewRouter(p)
	router.HandleFunc("/healthz", httputil.HealthCheckHandler)

	rootRouter = router.PathPrefix("/").Subrouter()
	rootRouter.Use((&deps.RequestMiddleware{
		RootProvider: p,
		ConfigSource: configSource,
	}).Handle)

	rootRouter.Use(p.Middleware(newSessionMiddleware))

	// TODO: move to another port
	session.ConfigureResolveHandler(rootRouter, p.Handler(newSessionResolveHandler))

	oauthRouter = rootRouter.NewRoute().Subrouter()
	oauthRouter.Use(p.Middleware(newCORSMiddleware))

	webappRouter = rootRouter.NewRoute().Subrouter()
	// When StrictSlash is true, the path in the browser URL always matches
	// the path specified in the route.
	// Trailing slash or missing slash will be corrected.
	// See http://www.gorillatoolkit.org/pkg/mux#Router.StrictSlash
	// Since our routes are specified without trailing slash,
	// the effect is that trailing slash is corrected with HTTP 301 by mux.
	webappRouter.StrictSlash(true)
	webappRouter.Use(webapp.IntlMiddleware)
	webappRouter.Use(p.Middleware(newCSPMiddleware))
	webappRouter.Use(p.Middleware(newCSRFMiddleware))
	webappRouter.Use(webapp.PostNoCacheMiddleware)
	webappRouter.Use(p.Middleware(newWebAppStateMiddleware))

	webappAuthRouter := webappRouter.NewRoute().Subrouter()
	webappAuthEntryPointRouter := webappAuthRouter.NewRoute().Subrouter()
	webappAuthEntryPointRouter.Use(p.Middleware(newAuthEntryPointMiddleware))
	webapphandler.ConfigureRootHandler(webappAuthEntryPointRouter, p.Handler(newWebAppRootHandler))
	webapphandler.ConfigureLoginHandler(webappAuthEntryPointRouter, p.Handler(newWebAppLoginHandler))
	webapphandler.ConfigureSignupHandler(webappAuthEntryPointRouter, p.Handler(newWebAppSignupHandler))
	webapphandler.ConfigurePromoteHandler(webappAuthEntryPointRouter, p.Handler(newWebAppPromoteHandler))

	webapphandler.ConfigureEnterPasswordHandler(webappAuthRouter, p.Handler(newWebAppEnterPasswordHandler))
	webapphandler.ConfigureEnterLoginIDHandler(webappAuthRouter, p.Handler(newWebAppEnterLoginIDHandler))
	webapphandler.ConfigureOOBOTPHandler(webappAuthRouter, p.Handler(newWebAppOOBOTPHandler))
	webapphandler.ConfigureCreatePasswordHandler(webappAuthRouter, p.Handler(newWebAppCreatePasswordHandler))
	webapphandler.ConfigureForgotPasswordHandler(webappAuthRouter, p.Handler(newWebAppForgotPasswordHandler))
	webapphandler.ConfigureForgotPasswordSuccessHandler(webappAuthRouter, p.Handler(newWebAppForgotPasswordSuccessHandler))
	webapphandler.ConfigureResetPasswordHandler(webappAuthRouter, p.Handler(newWebAppResetPasswordHandler))
	webapphandler.ConfigureResetPasswordSuccessHandler(webappAuthRouter, p.Handler(newWebAppResetPasswordSuccessHandler))

	webappAuthenticatedRouter := webappRouter.NewRoute().Subrouter()
	webappAuthenticatedRouter.Use(webapp.RequireAuthenticatedMiddleware{}.Handle)
	webapphandler.ConfigureSettingsHandler(webappAuthenticatedRouter, p.Handler(newWebAppSettingsHandler))
	webapphandler.ConfigureSettingsIdentityHandler(webappAuthenticatedRouter, p.Handler(newWebAppSettingsIdentityHandler))
	webapphandler.ConfigureLogoutHandler(webappAuthenticatedRouter, p.Handler(newWebAppLogoutHandler))

	webappSSOCallbackRouter := rootRouter.NewRoute().Subrouter()
	webappSSOCallbackRouter.Use(webapp.PostNoCacheMiddleware)
	webapphandler.ConfigureSSOCallbackHandler(webappSSOCallbackRouter, p.Handler(newWebAppSSOCallbackHandler))

	if p.ServerConfig.StaticAsset.ServingEnabled {
		fileServer := http.FileServer(http.Dir(p.ServerConfig.StaticAsset.Dir))
		rootRouter.PathPrefix("/static/").
			Handler(http.StripPrefix("/static/", fileServer))
	}

	oauthhandler.ConfigureMetadataHandler(oauthRouter, p.Handler(newOAuthMetadataHandler))
	oauthhandler.ConfigureJWKSHandler(oauthRouter, p.Handler(newOAuthJWKSHandler))
	oauthhandler.ConfigureAuthorizeHandler(oauthRouter, p.Handler(newOAuthAuthorizeHandler))
	oauthhandler.ConfigureTokenHandler(oauthRouter, p.Handler(newOAuthTokenHandler))
	oauthhandler.ConfigureRevokeHandler(oauthRouter, p.Handler(newOAuthRevokeHandler))
	oauthhandler.ConfigureUserInfoHandler(oauthRouter, p.Handler(newOAuthUserInfoHandler))
	oauthhandler.ConfigureEndSessionHandler(oauthRouter, p.Handler(newOAuthEndSessionHandler))
	oauthhandler.ConfigureChallengeHandler(oauthRouter, p.Handler(newOAuthChallengeHandler))

	return router
}
