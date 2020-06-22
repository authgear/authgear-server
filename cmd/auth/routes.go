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
	"github.com/skygeario/skygear-server/pkg/middlewares"
)

func NewRouter(p *deps.RootProvider) *mux.Router {
	rootRouter := mux.NewRouter()
	rootRouter.Use(sentry.Middleware(sentry.DefaultClient.Hub))
	rootRouter.Use(p.Middleware(middlewares.NewRecoverMiddleware))
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

	rootRouter.Use(p.Middleware(middlewares.NewSessionMiddleware))

	// TODO: move to another port
	session.AttachResolveHandler(rootRouter, p)

	oauthRouter = rootRouter.NewRoute().Subrouter()
	oauthRouter.Use(p.Middleware(middlewares.NewCORSMiddleware))

	webappRouter = rootRouter.NewRoute().Subrouter()
	// When StrictSlash is true, the path in the browser URL always matches
	// the path specified in the route.
	// Trailing slash or missing slash will be corrected.
	// See http://www.gorillatoolkit.org/pkg/mux#Router.StrictSlash
	// Since our routes are specified without trailing slash,
	// the effect is that trailing slash is corrected with HTTP 301 by mux.
	webappRouter.StrictSlash(true)
	webappRouter.Use(webapp.IntlMiddleware)
	webappRouter.Use(p.Middleware(middlewares.NewClientIDMiddleware))
	webappRouter.Use(p.Middleware(middlewares.NewCSPMiddleware))
	webappRouter.Use(p.Middleware(middlewares.NewCSRFMiddleware))
	webappRouter.Use(webapp.PostNoCacheMiddleware)
	webappRouter.Use(p.Middleware(middlewares.NewStateMiddleware))

	webappAuthRouter := webappRouter.NewRoute().Subrouter()
	webappAuthEntryPointRouter := webappAuthRouter.NewRoute().Subrouter()
	webappAuthEntryPointRouter.Use(webapp.AuthEntryPointMiddleware{}.Handle)
	webapphandler.AttachRootHandler(webappAuthEntryPointRouter, p)
	webapphandler.AttachLoginHandler(webappAuthEntryPointRouter, p)
	webapphandler.AttachSignupHandler(webappAuthEntryPointRouter, p)
	webapphandler.AttachPromoteHandler(webappAuthEntryPointRouter, p)

	webapphandler.AttachEnterPasswordHandler(webappAuthRouter, p)
	webapphandler.AttachEnterLoginIDHandler(webappAuthRouter, p)
	webapphandler.AttachOOBOTPHandler(webappAuthRouter, p)
	webapphandler.AttachCreatePasswordHandler(webappAuthRouter, p)
	webapphandler.AttachForgotPasswordHandler(webappAuthRouter, p)
	webapphandler.AttachForgotPasswordSuccessHandler(webappAuthRouter, p)
	webapphandler.AttachResetPasswordHandler(webappAuthRouter, p)
	webapphandler.AttachResetPasswordSuccessHandler(webappAuthRouter, p)

	webappAuthenticatedRouter := webappRouter.NewRoute().Subrouter()
	webappAuthenticatedRouter.Use(webapp.RequireAuthenticatedMiddleware{}.Handle)
	webapphandler.AttachSettingsHandler(webappAuthenticatedRouter, p)
	webapphandler.AttachSettingsIdentityHandler(webappAuthenticatedRouter, p)
	webapphandler.AttachLogoutHandler(webappAuthenticatedRouter, p)

	webappSSOCallbackRouter := rootRouter.NewRoute().Subrouter()
	webappSSOCallbackRouter.Use(webapp.PostNoCacheMiddleware)
	webapphandler.AttachSSOCallbackHandler(webappSSOCallbackRouter, p)

	if p.ServerConfig.StaticAsset.ServingEnabled {
		fileServer := http.FileServer(http.Dir(p.ServerConfig.StaticAsset.Dir))
		rootRouter.PathPrefix("/static/").
			Handler(http.StripPrefix("/static/", fileServer))
	}

	oauthhandler.AttachMetadataHandler(oauthRouter, p)
	oauthhandler.AttachJWKSHandler(oauthRouter, p)
	oauthhandler.AttachAuthorizeHandler(oauthRouter, p)
	oauthhandler.AttachTokenHandler(oauthRouter, p)
	oauthhandler.AttachRevokeHandler(oauthRouter, p)
	oauthhandler.AttachUserInfoHandler(oauthRouter, p)
	oauthhandler.AttachEndSessionHandler(oauthRouter, p)
	oauthhandler.AttachChallengeHandler(oauthRouter, p)

	return router
}
