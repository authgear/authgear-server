package main

import (
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
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
		ssohandler.CustomTokenLoginRequestSchema,
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

	var reservedNameChecker *password.ReservedNameChecker
	reservedNameChecker, err = password.NewReservedNameChecker(configuration.ReservedNameSourceFile)
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

	serverOption := server.DefaultOption()
	serverOption.GearPathPrefix = "/_auth"
	var srv server.Server
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

		srv = server.NewServerWithOption(configuration.Host, serverOption)
		srv.Router.Use(middleware.WriteTenantConfigMiddleware{
			ConfigurationProvider: middleware.ConfigurationProviderFunc(func(_ *http.Request) (config.TenantConfiguration, error) {
				return *tenantConfig, nil
			}),
		}.Handle)
		srv.Router.Use(middleware.ValidateHostMiddleware{ValidHosts: configuration.ValidHosts}.Handle)
		srv.Router.Use(middleware.RequestIDMiddleware{}.Handle)
		srv.Router.Use(middleware.CORSMiddleware{}.Handle)
	} else {
		srv = server.NewServerWithOption(configuration.Host, serverOption)
		srv.Router.Use(middleware.ReadTenantConfigMiddleware{}.Handle)
	}

	srv.Router.Use(middleware.DBMiddleware{Pool: dbPool}.Handle)
	srv.Router.Use(middleware.RedisMiddleware{Pool: redisPool}.Handle)
	srv.Router.Use(middleware.AuthMiddleware{}.Handle)

	srv.Router.Use(middleware.Injecter{
		MiddlewareFactory: middleware.AuthnMiddlewareFactory{},
		Dependency:        authDependency,
	}.Handle)

	handler.AttachSignupHandler(&srv, authDependency)
	handler.AttachLoginHandler(&srv, authDependency)
	handler.AttachLogoutHandler(&srv, authDependency)
	handler.AttachRefreshHandler(&srv, authDependency)
	handler.AttachMeHandler(&srv, authDependency)
	handler.AttachSetDisableHandler(&srv, authDependency)
	handler.AttachChangePasswordHandler(&srv, authDependency)
	handler.AttachResetPasswordHandler(&srv, authDependency)
	handler.AttachUpdateMetadataHandler(&srv, authDependency)
	handler.AttachListIdentitiesHandler(&srv, authDependency)
	forgotpwdhandler.AttachForgotPasswordHandler(&srv, authDependency)
	forgotpwdhandler.AttachForgotPasswordResetHandler(&srv, authDependency)
	userverifyhandler.AttachVerifyRequestHandler(&srv, authDependency)
	userverifyhandler.AttachVerifyCodeHandler(&srv, authDependency)
	userverifyhandler.AttachUpdateHandler(&srv, authDependency)
	ssohandler.AttachAuthURLHandler(&srv, authDependency)
	ssohandler.AttachAuthRedirectHandler(&srv, authDependency)
	ssohandler.AttachAuthHandler(&srv, authDependency)
	ssohandler.AttachAuthResultHandler(&srv, authDependency)
	ssohandler.AttachConfigHandler(&srv, authDependency)
	ssohandler.AttachCustomTokenLoginHandler(&srv, authDependency)
	ssohandler.AttachLoginHandler(&srv, authDependency)
	ssohandler.AttachLinkHandler(&srv, authDependency)
	ssohandler.AttachUnlinkHandler(&srv, authDependency)
	session.AttachListHandler(&srv, authDependency)
	session.AttachGetHandler(&srv, authDependency)
	session.AttachRevokeHandler(&srv, authDependency)
	session.AttachRevokeAllHandler(&srv, authDependency)
	mfaHandler.AttachListRecoveryCodeHandler(&srv, authDependency)
	mfaHandler.AttachRegenerateRecoveryCodeHandler(&srv, authDependency)
	mfaHandler.AttachListAuthenticatorHandler(&srv, authDependency)
	mfaHandler.AttachCreateTOTPHandler(&srv, authDependency)
	mfaHandler.AttachActivateTOTPHandler(&srv, authDependency)
	mfaHandler.AttachDeleteAuthenticatorHandler(&srv, authDependency)
	mfaHandler.AttachAuthenticateTOTPHandler(&srv, authDependency)
	mfaHandler.AttachRevokeAllBearerTokenHandler(&srv, authDependency)
	mfaHandler.AttachAuthenticateRecoveryCodeHandler(&srv, authDependency)
	mfaHandler.AttachAuthenticateBearerTokenHandler(&srv, authDependency)
	mfaHandler.AttachCreateOOBHandler(&srv, authDependency)
	mfaHandler.AttachTriggerOOBHandler(&srv, authDependency)
	mfaHandler.AttachActivateOOBHandler(&srv, authDependency)
	mfaHandler.AttachAuthenticateOOBHandler(&srv, authDependency)
	gearHandler.AttachTemplatesHandler(&srv, authDependency)
	loginidhandler.AttachAddLoginIDHandler(&srv, authDependency)
	loginidhandler.AttachRemoveLoginIDHandler(&srv, authDependency)
	loginidhandler.AttachUpdateLoginIDHandler(&srv, authDependency)

	server.ListenAndServe(srv.Server, logger, "Starting auth gear")
}
