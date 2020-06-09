package main

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/webapp"
	oauthhandler "github.com/skygeario/skygear-server/pkg/auth/handler/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/handler/session"
	webapphandler "github.com/skygeario/skygear-server/pkg/auth/handler/webapp"
	"github.com/skygeario/skygear-server/pkg/auth/task"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/middleware"
	"github.com/skygeario/skygear-server/pkg/core/redis"
	"github.com/skygeario/skygear-server/pkg/core/sentry"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/template"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

type configuration struct {
	Standalone                        bool
	StandaloneTenantConfigurationFile string                      `envconfig:"STANDALONE_TENANT_CONFIG_FILE" default:"standalone-tenant-config.yaml"`
	Host                              string                      `envconfig:"SERVER_HOST" default:"localhost:3000"`
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

/*
	@API Auth Gear
	@Version 1.0.0
	@Server {base_url}/_auth
		Auth Gear URL
		@Variable base_url https://my_app.skygearapis.com
			Skygear App URL

	@SecuritySchemeAPIKey access_key header X-Skygear-API-Key
		Access key used by client app
	@SecuritySchemeAPIKey master_key header X-Skygear-API-Key
		Master key used by admins, can perform administrative operations.
		Can be used as access key as well.
	@SecuritySchemeHTTP access_token Bearer token
		Access token of user
	@SecurityRequirement access_key

	@Tag User
		User information
	@Tag Forgot Password
		Password recovery process
	@Tag Administration
		Administrative operation
	@Tag SSO
		Single sign-on
*/
func main() {
	// logging initialization
	logging.SetModule("auth")
	loggerFactory := logging.NewFactory(
		logging.NewDefaultLogHook(nil),
		&sentry.LogHook{Hub: sentry.DefaultClient.Hub},
	)
	logger := loggerFactory.NewLogger("auth")

	envErr := godotenv.Load()
	if envErr != nil {
		logger.WithError(envErr).Debug("Cannot load .env file")
	}

	configuration := configuration{}
	envconfig.Process("", &configuration)
	if configuration.ValidHosts == "" {
		configuration.ValidHosts = configuration.Host
	}

	validator := validation.NewValidator("http://v2.skgyear.io")
	validator.AddSchemaFragments(
		oauthhandler.ChallengeRequestSchema,
	)

	dbPool := db.NewPool()
	redisPool, err := redis.NewPool(configuration.Redis)
	if err != nil {
		logger.Fatalf("fail to create redis pool: %v", err.Error())
	}
	asyncTaskExecutor := async.NewExecutor(dbPool)
	var assetGearLoader *template.AssetGearLoader
	if configuration.Template.AssetGearEndpoint != "" && configuration.Template.AssetGearMasterKey != "" {
		assetGearLoader = &template.AssetGearLoader{
			AssetGearEndpoint:  configuration.Template.AssetGearEndpoint,
			AssetGearMasterKey: configuration.Template.AssetGearMasterKey,
		}
	}

	var reservedNameChecker *loginid.ReservedNameChecker
	reservedNameChecker, err = loginid.NewReservedNameChecker(configuration.ReservedNameSourceFile)
	if err != nil {
		logger.Fatalf("fail to load reserved name source file: %v", err.Error())
	}

	authDependency := auth.DependencyMap{
		EnableFileSystemTemplate: configuration.Template.EnableFileLoader,
		AssetGearLoader:          assetGearLoader,
		AsyncTaskExecutor:        asyncTaskExecutor,
		UseInsecureCookie:        configuration.UseInsecureCookie,
		StaticAssetURLPrefix:     configuration.StaticAssetURLPrefix,
		DefaultConfiguration:     configuration.Default,
		Validator:                validator,
		ReservedNameChecker:      reservedNameChecker,
	}

	task.AttachPwHousekeeperTask(asyncTaskExecutor, authDependency)
	task.AttachSendMessagesTask(asyncTaskExecutor, authDependency)

	var router *mux.Router
	var rootRouter *mux.Router
	var webappRouter *mux.Router
	var oauthRouter *mux.Router
	if configuration.Standalone {
		filename := configuration.StandaloneTenantConfigurationFile
		reader, err := os.Open(filename)
		if err != nil {
			logger.WithError(err).Error("Cannot open standalone config")
		}
		tenantConfig, err := config.NewTenantConfigurationFromYAML(reader)
		if err != nil {
			logger.WithError(err).Fatal("Cannot parse standalone config")
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

	rootRouter.Use(middleware.DBMiddleware{Pool: dbPool}.Handle)
	rootRouter.Use(middleware.RedisMiddleware{Pool: redisPool}.Handle)
	rootRouter.Use(auth.MakeMiddleware(authDependency, auth.NewSessionMiddleware))
	// The resolve endpoint is now mounted at root router.
	// Therefore the access key middleware needs to be mounted at root router as well.
	rootRouter.Use(auth.MakeMiddleware(authDependency, auth.NewAccessKeyMiddleware))

	if configuration.Standalone {
		// Attach resolve endpoint in the router that does not validate host.
		session.AttachResolveHandler(rootRouter, authDependency)
		rootRouter = rootRouter.NewRoute().Subrouter()
		rootRouter.Use(middleware.ValidateHostMiddleware{ValidHosts: configuration.ValidHosts}.Handle)

		oauthRouter = rootRouter.NewRoute().Subrouter()
		oauthRouter.Use(middleware.CORSMiddleware{}.Handle)
	} else {
		session.AttachResolveHandler(rootRouter, authDependency)

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
	webappRouter.Use(auth.MakeMiddleware(authDependency, auth.NewClientIDMiddleware))
	webappRouter.Use(auth.MakeMiddleware(authDependency, auth.NewCSPMiddleware))
	webappRouter.Use(auth.MakeMiddleware(authDependency, auth.NewCSRFMiddleware))
	webappRouter.Use(webapp.PostNoCacheMiddleware)
	webappRouter.Use(auth.MakeMiddleware(authDependency, auth.NewStateMiddleware))

	webappAuthRouter := webappRouter.NewRoute().Subrouter()
	webappAuthEntryPointRouter := webappAuthRouter.NewRoute().Subrouter()
	webappAuthEntryPointRouter.Use(webapp.AuthEntryPointMiddleware{}.Handle)
	webapphandler.AttachRootHandler(webappAuthEntryPointRouter, authDependency)
	webapphandler.AttachLoginHandler(webappAuthEntryPointRouter, authDependency)
	webapphandler.AttachSignupHandler(webappAuthEntryPointRouter, authDependency)
	webapphandler.AttachPromoteHandler(webappAuthEntryPointRouter, authDependency)

	webapphandler.AttachEnterPasswordHandler(webappAuthRouter, authDependency)
	webapphandler.AttachEnterLoginIDHandler(webappAuthRouter, authDependency)
	webapphandler.AttachOOBOTPHandler(webappAuthRouter, authDependency)
	webapphandler.AttachCreatePasswordHandler(webappAuthRouter, authDependency)
	webapphandler.AttachForgotPasswordHandler(webappAuthRouter, authDependency)
	webapphandler.AttachForgotPasswordSuccessHandler(webappAuthRouter, authDependency)
	webapphandler.AttachResetPasswordHandler(webappAuthRouter, authDependency)
	webapphandler.AttachResetPasswordSuccessHandler(webappAuthRouter, authDependency)

	webappAuthenticatedRouter := webappRouter.NewRoute().Subrouter()
	webappAuthenticatedRouter.Use(webapp.RequireAuthenticatedMiddleware{}.Handle)
	webapphandler.AttachSettingsHandler(webappAuthenticatedRouter, authDependency)
	webapphandler.AttachSettingsIdentityHandler(webappAuthenticatedRouter, authDependency)
	webapphandler.AttachLogoutHandler(webappAuthenticatedRouter, authDependency)

	webappSSOCallbackRouter := rootRouter.NewRoute().Subrouter()
	webappSSOCallbackRouter.Use(webapp.PostNoCacheMiddleware)
	webapphandler.AttachSSOCallbackHandler(webappSSOCallbackRouter, authDependency)

	if configuration.StaticAssetDir != "" {
		rootRouter.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(configuration.StaticAssetDir))))
	}

	oauthhandler.AttachMetadataHandler(oauthRouter, authDependency)
	oauthhandler.AttachJWKSHandler(oauthRouter, authDependency)
	oauthhandler.AttachAuthorizeHandler(oauthRouter, authDependency)
	oauthhandler.AttachTokenHandler(oauthRouter, authDependency)
	oauthhandler.AttachRevokeHandler(oauthRouter, authDependency)
	oauthhandler.AttachUserInfoHandler(oauthRouter, authDependency)
	oauthhandler.AttachEndSessionHandler(oauthRouter, authDependency)
	oauthhandler.AttachChallengeHandler(oauthRouter, authDependency)

	srv := &http.Server{
		Addr:    configuration.Host,
		Handler: router,
	}
	server.ListenAndServe(srv, logger, "Starting auth gear")
}
