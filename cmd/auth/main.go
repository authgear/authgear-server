package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/redis"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"

	"github.com/skygeario/skygear-server/pkg/auth"
	authTemplate "github.com/skygeario/skygear-server/pkg/auth/template"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/template"

	"github.com/skygeario/skygear-server/pkg/auth/handler"
	forgotpwdhandler "github.com/skygeario/skygear-server/pkg/auth/handler/forgotpwd"
	mfaHandler "github.com/skygeario/skygear-server/pkg/auth/handler/mfa"
	"github.com/skygeario/skygear-server/pkg/auth/handler/session"
	ssohandler "github.com/skygeario/skygear-server/pkg/auth/handler/sso"
	userverifyhandler "github.com/skygeario/skygear-server/pkg/auth/handler/userverify"
	"github.com/skygeario/skygear-server/pkg/auth/task"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/middleware"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

type configuration struct {
	Standalone                        bool
	StandaloneTenantConfigurationFile string                      `envconfig:"STANDALONE_TENANT_CONFIG_FILE" default:"standalone-tenant-config.yaml"`
	PathPrefix                        string                      `envconfig:"PATH_PREFIX"`
	Host                              string                      `envconfig:"SERVER_HOST" default:"localhost:3000"`
	ValidHosts                        string                      `envconfig:"VALID_HOSTS"`
	Redis                             redis.Configuration         `envconfig:"REDIS"`
	UseInsecureCookie                 bool                        `envconfig:"INSECURE_COOKIE"`
	Default                           config.DefaultConfiguration `envconfig:"DEFAULT"`
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
	@SecuritySchemeHTTP access_token Bearer JWT
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
	loggerFactory := logging.NewFactory(logging.NewDefaultMaskedTextFormatter(nil))
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
	if configuration.Redis.Host == "" {
		logger.Fatal("REDIS_HOST is not provided")
	}

	// default template initialization
	templateEngine := template.NewEngine()
	authTemplate.RegisterDefaultTemplates(templateEngine)

	dbPool := db.NewPool()
	redisPool := redis.NewPool(configuration.Redis)
	asyncTaskExecutor := async.NewExecutor(dbPool)
	authDependency := auth.DependencyMap{
		AsyncTaskExecutor:    asyncTaskExecutor,
		TemplateEngine:       templateEngine,
		UseInsecureCookie:    configuration.UseInsecureCookie,
		DefaultConfiguration: configuration.Default,
	}

	task.AttachVerifyCodeSendTask(asyncTaskExecutor, authDependency)
	task.AttachPwHousekeeperTask(asyncTaskExecutor, authDependency)
	task.AttachWelcomeEmailSendTask(asyncTaskExecutor, authDependency)

	var srv server.Server
	if configuration.Standalone {
		filename := configuration.StandaloneTenantConfigurationFile
		reader, err := os.Open(filename)
		if err != nil {
			logger.WithError(err).Error("Cannot open standalone config")
		}
		tenantConfig, err := config.NewTenantConfigurationFromYAML(reader)
		if err != nil {
			if skyError, ok := err.(skyerr.Error); ok {
				info := skyError.Info()
				if arguments, ok := info["arguments"].([]string); ok {
					for _, a := range arguments {
						fmt.Fprintf(os.Stderr, "%v\n", a)
					}
				}
			}
			logger.WithError(err).Fatal("Cannot parse standalone config")
		}

		serverOption := server.DefaultOption()
		serverOption.GearPathPrefix = configuration.PathPrefix
		srv = server.NewServerWithOption(configuration.Host, authDependency, serverOption)
		srv.Use(middleware.TenantConfigurationMiddleware{
			ConfigurationProvider: middleware.ConfigurationProviderFunc(func(_ *http.Request) (config.TenantConfiguration, error) {
				return *tenantConfig, nil
			}),
		}.Handle)
		srv.Use(middleware.ValidateHostMiddleware{ValidHosts: configuration.ValidHosts}.Handle)
		srv.Use(middleware.RequestIDMiddleware{}.Handle)
		srv.Use(middleware.CORSMiddleware{}.Handle)
	} else {
		srv = server.NewServer(configuration.Host, authDependency)
	}

	srv.Use(middleware.DBMiddleware{Pool: dbPool}.Handle)
	srv.Use(middleware.RedisMiddleware{Pool: redisPool}.Handle)
	srv.Use(middleware.AuthMiddleware{}.Handle)

	srv.Use(middleware.Injecter{
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
	handler.AttachWelcomeEmailHandler(&srv, authDependency)
	handler.AttachUpdateMetadataHandler(&srv, authDependency)
	forgotpwdhandler.AttachForgotPasswordHandler(&srv, authDependency)
	forgotpwdhandler.AttachForgotPasswordResetHandler(&srv, authDependency)
	userverifyhandler.AttachVerifyRequestHandler(&srv, authDependency)
	userverifyhandler.AttachVerifyCodeHandler(&srv, authDependency)
	ssohandler.AttachAuthURLHandler(&srv, authDependency)
	ssohandler.AttachConfigHandler(&srv, authDependency)
	ssohandler.AttachIFrameHandlerFactory(&srv, authDependency)
	ssohandler.AttachCustomTokenLoginHandler(&srv, authDependency)
	ssohandler.AttachAuthHandler(&srv, authDependency)
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
	mfaHandler.AttachTOTPQRCodeHandler(&srv, authDependency)

	go func() {
		logger.Info("Starting auth gear")
		if err := srv.ListenAndServe(); err != nil {
			logger.WithError(err).Error("Cannot start HTTP server")
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	// wait interrupt signal
	select {
	case <-sig:
		logger.Info("Stopping HTTP server")
	}

	// create shutdown context with 10 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// shutdown the server
	err := srv.Shutdown(ctx)
	if err != nil {
		logger.WithError(err).Fatal("Cannot shutdown server")
	}
}
