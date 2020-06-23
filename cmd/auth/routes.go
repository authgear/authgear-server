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
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/deps"
)

func NewRouter(p *deps.RootProvider) *mux.Router {
	rootRouter := mux.NewRouter()
	rootRouter.Use(p.Middleware(newSentryMiddlewareFactory(sentry.DefaultClient.Hub)))
	rootRouter.Use(p.Middleware(newRecoverMiddleware))
	return rootRouter
}

func setupRoutes(p *deps.RootProvider, configSource configsource.Source) *mux.Router {
	var router *mux.Router
	var rootRouter *mux.Router
	var webappRouter *mux.Router
	var oauthRouter *mux.Router

	router = NewRouter(p)
	router.HandleFunc("/healthz", server.HealthCheckHandler)

	rootRouter = router.PathPrefix("/").Subrouter()
	rootRouter.Use((&deps.RequestMiddleware{
		RootProvider: p,
		ConfigSource: configSource,
	}).Handle)

	rootRouter.Use(p.Middleware(nil)) // auth.Middleware

	// TODO: move to another port
	session.ConfigureResolveHandler(rootRouter, nil)

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
	webappRouter.Use(p.Middleware(nil)) // webapp.StateMiddleware

	webappAuthRouter := webappRouter.NewRoute().Subrouter()
	webappAuthEntryPointRouter := webappAuthRouter.NewRoute().Subrouter()
	webappAuthEntryPointRouter.Use(p.Middleware(newAuthEntryPointMiddleware))
	webapphandler.ConfigureRootHandler(webappAuthEntryPointRouter, nil)
	webapphandler.ConfigureLoginHandler(webappAuthEntryPointRouter, nil)
	webapphandler.ConfigureSignupHandler(webappAuthEntryPointRouter, nil)
	webapphandler.ConfigurePromoteHandler(webappAuthEntryPointRouter, nil)

	webapphandler.ConfigureEnterPasswordHandler(webappAuthRouter, nil)
	webapphandler.ConfigureEnterLoginIDHandler(webappAuthRouter, nil)
	webapphandler.ConfigureOOBOTPHandler(webappAuthRouter, nil)
	webapphandler.ConfigureCreatePasswordHandler(webappAuthRouter, nil)
	webapphandler.ConfigureForgotPasswordHandler(webappAuthRouter, nil)
	webapphandler.ConfigureForgotPasswordSuccessHandler(webappAuthRouter, nil)
	webapphandler.ConfigureResetPasswordHandler(webappAuthRouter, nil)
	webapphandler.ConfigureResetPasswordSuccessHandler(webappAuthRouter, nil)

	webappAuthenticatedRouter := webappRouter.NewRoute().Subrouter()
	webappAuthenticatedRouter.Use(webapp.RequireAuthenticatedMiddleware{}.Handle)
	webapphandler.ConfigureSettingsHandler(webappAuthenticatedRouter, nil)
	webapphandler.ConfigureSettingsIdentityHandler(webappAuthenticatedRouter, nil)
	webapphandler.ConfigureLogoutHandler(webappAuthenticatedRouter, nil)

	webappSSOCallbackRouter := rootRouter.NewRoute().Subrouter()
	webappSSOCallbackRouter.Use(webapp.PostNoCacheMiddleware)
	webapphandler.ConfigureSSOCallbackHandler(webappSSOCallbackRouter, nil)

	if p.ServerConfig.StaticAsset.ServingEnabled {
		fileServer := http.FileServer(http.Dir(p.ServerConfig.StaticAsset.Dir))
		rootRouter.PathPrefix("/static/").
			Handler(http.StripPrefix("/static/", fileServer))
	}

	oauthhandler.ConfigureMetadataHandler(oauthRouter, nil)
	oauthhandler.ConfigureJWKSHandler(oauthRouter, nil)
	oauthhandler.ConfigureAuthorizeHandler(oauthRouter, nil)
	oauthhandler.ConfigureTokenHandler(oauthRouter, nil)
	oauthhandler.ConfigureRevokeHandler(oauthRouter, nil)
	oauthhandler.ConfigureUserInfoHandler(oauthRouter, nil)
	oauthhandler.ConfigureEndSessionHandler(oauthRouter, nil)
	oauthhandler.ConfigureChallengeHandler(oauthRouter, nil)

	return router
}
