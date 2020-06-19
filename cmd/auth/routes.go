package main

import (
	"log"
	"net/http"
	"os"

	redigo "github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"

	configsource "github.com/skygeario/skygear-server/pkg/auth/config/source"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/webapp"
	oauthhandler "github.com/skygeario/skygear-server/pkg/auth/handler/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/handler/session"
	webapphandler "github.com/skygeario/skygear-server/pkg/auth/handler/webapp"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/middleware"
	"github.com/skygeario/skygear-server/pkg/core/redis"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/deps"
	"github.com/skygeario/skygear-server/pkg/middlewares"
)

type configuration struct {
	Standalone                        bool
	StandaloneTenantConfigurationFile string                      `envconfig:"STANDALONE_TENANT_CONFIG_FILE" default:"standalone-tenant-config.yaml"`
	Host                              string                      `envconfig:"SERVER_HOST" default:"0.0.0.0:3000"`
	ValidHosts                        string                      `envconfig:"VALID_HOSTS"`
	Redis                             redis.Configuration         `envconfig:"REDIS"`
	UseInsecureCookie                 bool                        `envconfig:"INSECURE_COOKIE"`
	Template                          TemplateConfiguration       `envconfig:"TEMPLATE"`
	Default                           config.DefaultConfiguration `envconfig:"DEFAULT"`
	ReservedNameSourceFile            string                      `envconfig:"RESERVED_NAME_SOURCE_FILE" default:"reserved_name.txt"`
	// StaticAssetDir is for serving the static asset locally.
	// It should not be used for production.
	StaticAssetDir string `envconfig:"STATIC_ASSET_DIR"`
	// StaticAssetURLPrefix sets the prefix for static asset.
	// In production, it should look like https://code.skygear.dev/dist/git-<commit-hash>/authui
	StaticAssetURLPrefix string `envconfig:"STATIC_ASSET_URL_PREFIX"`
}

type TemplateConfiguration struct {
	EnableFileLoader   bool   `envconfig:"ENABLE_FILE_LOADER"`
	AssetGearEndpoint  string `envconfig:"ASSET_GEAR_ENDPOINT"`
	AssetGearMasterKey string `envconfig:"ASSET_GEAR_MASTER_KEY"`
}

func setupNewRoutes(p *deps.RootProvider, configSource configsource.Source) *mux.Router {
	router := server.NewRouter()
	router.HandleFunc("/healthz", server.HealthCheckHandler)

	rootRouter := router.PathPrefix("/").Subrouter()
	rootRouter.Use((&deps.RequestMiddleware{
		RootProvider: p,
		ConfigSource: configSource,
	}).Handle)

	return router
}

// nolint: deadcode
func setupRoutes(cfg configuration, redisPool *redigo.Pool, p *deps.RootProvider) *mux.Router {
	var router *mux.Router
	var rootRouter *mux.Router
	var webappRouter *mux.Router
	var oauthRouter *mux.Router
	if cfg.Standalone {
		filename := cfg.StandaloneTenantConfigurationFile
		reader, err := os.Open(filename)
		if err != nil {
			log.Fatal("Cannot open standalone config")
		}
		tenantConfig, err := config.NewTenantConfigurationFromYAML(reader)
		if err != nil {
			log.Fatal("Cannot parse standalone config")
		}

		router = server.NewRouter()
		router.HandleFunc("/healthz", server.HealthCheckHandler)

		rootRouter = router.PathPrefix("/").Subrouter()
		rootRouter.Use(middleware.WriteTenantConfigMiddleware{
			ConfigurationProvider: middleware.ConfigurationProviderFunc(func(_ *http.Request) (config.TenantConfiguration, error) {
				return *tenantConfig, nil
			}),
		}.Handle)
	} else {
		router = server.NewRouter()
		router.HandleFunc("/healthz", server.HealthCheckHandler)

		rootRouter = router.PathPrefix("/").Subrouter()
		rootRouter.Use(middleware.ReadTenantConfigMiddleware{}.Handle)
	}

	rootRouter.Use(middleware.RedisMiddleware{Pool: redisPool}.Handle)
	rootRouter.Use(p.Middleware(middlewares.NewSessionMiddleware))

	if cfg.Standalone {
		// Attach resolve endpoint in the router that does not validate host.
		session.AttachResolveHandler(rootRouter, p)
		rootRouter = rootRouter.NewRoute().Subrouter()
		rootRouter.Use(middleware.ValidateHostMiddleware{ValidHosts: cfg.ValidHosts}.Handle)

		oauthRouter = rootRouter.NewRoute().Subrouter()
		oauthRouter.Use(middleware.CORSMiddleware{}.Handle)
	} else {
		session.AttachResolveHandler(rootRouter, p)

		oauthRouter = rootRouter.NewRoute().Subrouter()
	}

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

	if cfg.StaticAssetDir != "" {
		rootRouter.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(cfg.StaticAssetDir))))
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
