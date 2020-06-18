package main

import (
	"net/http"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/loginid"
	oauthhandler "github.com/skygeario/skygear-server/pkg/auth/handler/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/task"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/redis"
	"github.com/skygeario/skygear-server/pkg/core/sentry"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/template"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

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

	router := setupRoutes(configuration, dbPool, redisPool, authDependency)

	srv := &http.Server{
		Addr:    configuration.Host,
		Handler: router,
	}
	server.ListenAndServe(srv, logger, "Starting auth gear")
}
