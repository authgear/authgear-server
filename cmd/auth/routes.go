package main

import (
	"log"
	"net/http"
	"os"

	redigo "github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/webapp"
	oauthhandler "github.com/skygeario/skygear-server/pkg/auth/handler/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/handler/session"
	webapphandler "github.com/skygeario/skygear-server/pkg/auth/handler/webapp"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/middleware"
	"github.com/skygeario/skygear-server/pkg/core/redis"
	"github.com/skygeario/skygear-server/pkg/core/server"
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

// nolint: deadcode
func setupRoutes(cfg configuration, redisPool *redigo.Pool, deps auth.DependencyMap) *mux.Router {
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
	rootRouter.Use(auth.MakeMiddleware(deps, auth.NewSessionMiddleware))
	// The resolve endpoint is now mounted at root router.
	// Therefore the access key middleware needs to be mounted at root router as well.
	rootRouter.Use(auth.MakeMiddleware(deps, auth.NewAccessKeyMiddleware))

	if cfg.Standalone {
		// Attach resolve endpoint in the router that does not validate host.
		session.AttachResolveHandler(rootRouter, deps)
		rootRouter = rootRouter.NewRoute().Subrouter()
		rootRouter.Use(middleware.ValidateHostMiddleware{ValidHosts: cfg.ValidHosts}.Handle)

		oauthRouter = rootRouter.NewRoute().Subrouter()
		oauthRouter.Use(middleware.CORSMiddleware{}.Handle)
	} else {
		session.AttachResolveHandler(rootRouter, deps)

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
	webappRouter.Use(auth.MakeMiddleware(deps, auth.NewClientIDMiddleware))
	webappRouter.Use(auth.MakeMiddleware(deps, auth.NewCSPMiddleware))
	webappRouter.Use(auth.MakeMiddleware(deps, auth.NewCSRFMiddleware))
	webappRouter.Use(webapp.PostNoCacheMiddleware)
	webappRouter.Use(auth.MakeMiddleware(deps, auth.NewStateMiddleware))

	webappAuthRouter := webappRouter.NewRoute().Subrouter()
	webappAuthEntryPointRouter := webappAuthRouter.NewRoute().Subrouter()
	webappAuthEntryPointRouter.Use(webapp.AuthEntryPointMiddleware{}.Handle)
	webapphandler.AttachRootHandler(webappAuthEntryPointRouter, deps)
	webapphandler.AttachLoginHandler(webappAuthEntryPointRouter, deps)
	webapphandler.AttachSignupHandler(webappAuthEntryPointRouter, deps)
	webapphandler.AttachPromoteHandler(webappAuthEntryPointRouter, deps)

	webapphandler.AttachEnterPasswordHandler(webappAuthRouter, deps)
	webapphandler.AttachEnterLoginIDHandler(webappAuthRouter, deps)
	webapphandler.AttachOOBOTPHandler(webappAuthRouter, deps)
	webapphandler.AttachCreatePasswordHandler(webappAuthRouter, deps)
	webapphandler.AttachForgotPasswordHandler(webappAuthRouter, deps)
	webapphandler.AttachForgotPasswordSuccessHandler(webappAuthRouter, deps)
	webapphandler.AttachResetPasswordHandler(webappAuthRouter, deps)
	webapphandler.AttachResetPasswordSuccessHandler(webappAuthRouter, deps)

	webappAuthenticatedRouter := webappRouter.NewRoute().Subrouter()
	webappAuthenticatedRouter.Use(webapp.RequireAuthenticatedMiddleware{}.Handle)
	webapphandler.AttachSettingsHandler(webappAuthenticatedRouter, deps)
	webapphandler.AttachSettingsIdentityHandler(webappAuthenticatedRouter, deps)
	webapphandler.AttachLogoutHandler(webappAuthenticatedRouter, deps)

	webappSSOCallbackRouter := rootRouter.NewRoute().Subrouter()
	webappSSOCallbackRouter.Use(webapp.PostNoCacheMiddleware)
	webapphandler.AttachSSOCallbackHandler(webappSSOCallbackRouter, deps)

	if cfg.StaticAssetDir != "" {
		rootRouter.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(cfg.StaticAssetDir))))
	}

	oauthhandler.AttachMetadataHandler(oauthRouter, deps)
	oauthhandler.AttachJWKSHandler(oauthRouter, deps)
	oauthhandler.AttachAuthorizeHandler(oauthRouter, deps)
	oauthhandler.AttachTokenHandler(oauthRouter, deps)
	oauthhandler.AttachRevokeHandler(oauthRouter, deps)
	oauthhandler.AttachUserInfoHandler(oauthRouter, deps)
	oauthhandler.AttachEndSessionHandler(oauthRouter, deps)
	oauthhandler.AttachChallengeHandler(oauthRouter, deps)

	return router
}
