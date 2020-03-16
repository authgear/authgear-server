package main

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/handler"
	forgotpwdhandler "github.com/skygeario/skygear-server/pkg/auth/handler/forgotpwd"
	gearHandler "github.com/skygeario/skygear-server/pkg/auth/handler/gear"
	loginidhandler "github.com/skygeario/skygear-server/pkg/auth/handler/loginid"
	mfaHandler "github.com/skygeario/skygear-server/pkg/auth/handler/mfa"
	"github.com/skygeario/skygear-server/pkg/auth/handler/session"
	ssohandler "github.com/skygeario/skygear-server/pkg/auth/handler/sso"
	userverifyhandler "github.com/skygeario/skygear-server/pkg/auth/handler/userverify"
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
	@Tag User Verification
		Login IDs verification
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
		handler.ChangePasswordRequestSchema,
		handler.SetDisableRequestSchema,
		handler.RefreshRequestSchema,
		handler.ResetPasswordRequestSchema,
		handler.LoginRequestSchema,
		handler.SignupRequestSchema,
		handler.UpdateMetadataRequestSchema,

		forgotpwdhandler.ForgotPasswordRequestSchema,
		forgotpwdhandler.ForgotPasswordResetPageSchema,
		forgotpwdhandler.ForgotPasswordResetFormSchema,
		forgotpwdhandler.ForgotPasswordResetRequestSchema,

		mfaHandler.ActivateOOBRequestSchema,
		mfaHandler.ActivateTOTPRequestSchema,
		mfaHandler.AuthenticateBearerTokenRequestSchema,
		mfaHandler.AuthenticateOOBRequestSchema,
		mfaHandler.AuthenticateRecoveryCodeRequestSchema,
		mfaHandler.AuthenticateTOTPRequestSchema,
		mfaHandler.CreateOOBRequestSchema,
		mfaHandler.CreateTOTPRequestSchema,
		mfaHandler.DeleteAuthenticatorRequestSchema,
		mfaHandler.ListAuthenticatorRequestSchema,
		mfaHandler.TriggerOOBRequestSchema,

		session.GetRequestSchema,
		session.RevokeRequestSchema,

		ssohandler.AuthURLRequestSchema,
		ssohandler.LoginRequestSchema,
		ssohandler.LinkRequestSchema,
		ssohandler.AuthResultRequestSchema,

		userverifyhandler.VerifyCodeRequestSchema,
		userverifyhandler.VerifyRequestSchema,
		userverifyhandler.VerifyCodeFormSchema,
		userverifyhandler.UpdateVerifyStateRequestSchema,

		loginidhandler.AddLoginIDRequestSchema,
		loginidhandler.RemoveLoginIDRequestSchema,
		loginidhandler.UpdateLoginIDRequestSchema,
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
		DefaultConfiguration:     configuration.Default,
		Validator:                validator,
		ReservedNameChecker:      reservedNameChecker,
	}

	task.AttachVerifyCodeSendTask(asyncTaskExecutor, authDependency)
	task.AttachPwHousekeeperTask(asyncTaskExecutor, authDependency)
	task.AttachWelcomeEmailSendTask(asyncTaskExecutor, authDependency)

	serverOption := server.Option{
		GearPathPrefix: "/_auth",
	}
	var rootRouter *mux.Router
	var appRouter *mux.Router
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

		rootRouter, appRouter = server.NewRouterWithOption(serverOption)
		appRouter.Use(middleware.WriteTenantConfigMiddleware{
			ConfigurationProvider: middleware.ConfigurationProviderFunc(func(_ *http.Request) (config.TenantConfiguration, error) {
				return *tenantConfig, nil
			}),
		}.Handle)
		appRouter.Use(middleware.ValidateHostMiddleware{ValidHosts: configuration.ValidHosts}.Handle)
		appRouter.Use(middleware.RequestIDMiddleware{}.Handle)
		appRouter.Use(middleware.CORSMiddleware{}.Handle)
	} else {
		rootRouter, appRouter = server.NewRouterWithOption(serverOption)
		appRouter.Use(middleware.ReadTenantConfigMiddleware{}.Handle)
	}

	appRouter.Use(middleware.DBMiddleware{Pool: dbPool}.Handle)
	appRouter.Use(middleware.RedisMiddleware{Pool: redisPool}.Handle)
	appRouter.Use(middleware.AuthMiddleware{}.Handle)

	appRouter.Use(middleware.Injecter{
		MiddlewareFactory: middleware.AuthnMiddlewareFactory{},
		Dependency:        authDependency,
	}.Handle)

	handler.AttachSignupHandler(appRouter, authDependency)
	handler.AttachLoginHandler(appRouter, authDependency)
	handler.AttachLogoutHandler(appRouter, authDependency)
	handler.AttachRefreshHandler(appRouter, authDependency)
	handler.AttachMeHandler(appRouter, authDependency)
	handler.AttachSetDisableHandler(appRouter, authDependency)
	handler.AttachChangePasswordHandler(appRouter, authDependency)
	handler.AttachResetPasswordHandler(appRouter, authDependency)
	handler.AttachUpdateMetadataHandler(appRouter, authDependency)
	handler.AttachListIdentitiesHandler(appRouter, authDependency)
	forgotpwdhandler.AttachForgotPasswordHandler(appRouter, authDependency)
	forgotpwdhandler.AttachForgotPasswordResetHandler(appRouter, authDependency)
	userverifyhandler.AttachVerifyRequestHandler(appRouter, authDependency)
	userverifyhandler.AttachVerifyCodeHandler(appRouter, authDependency)
	userverifyhandler.AttachUpdateHandler(appRouter, authDependency)
	ssohandler.AttachAuthURLHandler(appRouter, authDependency)
	ssohandler.AttachAuthRedirectHandler(appRouter, authDependency)
	ssohandler.AttachAuthHandler(appRouter, authDependency)
	ssohandler.AttachAuthResultHandler(appRouter, authDependency)
	ssohandler.AttachLoginHandler(appRouter, authDependency)
	ssohandler.AttachLinkHandler(appRouter, authDependency)
	ssohandler.AttachUnlinkHandler(appRouter, authDependency)
	session.AttachListHandler(appRouter, authDependency)
	session.AttachGetHandler(appRouter, authDependency)
	session.AttachRevokeHandler(appRouter, authDependency)
	session.AttachRevokeAllHandler(appRouter, authDependency)
	session.AttachResolveHandler(appRouter, authDependency)
	mfaHandler.AttachListRecoveryCodeHandler(appRouter, authDependency)
	mfaHandler.AttachRegenerateRecoveryCodeHandler(appRouter, authDependency)
	mfaHandler.AttachListAuthenticatorHandler(appRouter, authDependency)
	mfaHandler.AttachCreateTOTPHandler(appRouter, authDependency)
	mfaHandler.AttachActivateTOTPHandler(appRouter, authDependency)
	mfaHandler.AttachDeleteAuthenticatorHandler(appRouter, authDependency)
	mfaHandler.AttachAuthenticateTOTPHandler(appRouter, authDependency)
	mfaHandler.AttachRevokeAllBearerTokenHandler(appRouter, authDependency)
	mfaHandler.AttachAuthenticateRecoveryCodeHandler(appRouter, authDependency)
	mfaHandler.AttachAuthenticateBearerTokenHandler(appRouter, authDependency)
	mfaHandler.AttachCreateOOBHandler(appRouter, authDependency)
	mfaHandler.AttachTriggerOOBHandler(appRouter, authDependency)
	mfaHandler.AttachActivateOOBHandler(appRouter, authDependency)
	mfaHandler.AttachAuthenticateOOBHandler(appRouter, authDependency)
	gearHandler.AttachTemplatesHandler(appRouter, authDependency)
	loginidhandler.AttachAddLoginIDHandler(appRouter, authDependency)
	loginidhandler.AttachRemoveLoginIDHandler(appRouter, authDependency)
	loginidhandler.AttachUpdateLoginIDHandler(appRouter, authDependency)

	srv := &http.Server{
		Addr:    configuration.Host,
		Handler: rootRouter,
	}
	server.ListenAndServe(srv, logger, "Starting auth gear")
}
