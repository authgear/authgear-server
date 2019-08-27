package main

import (
	"context"
	"fmt"
	"log"
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
	"github.com/skygeario/skygear-server/pkg/core/auth/authn/resolver"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/template"

	"github.com/skygeario/skygear-server/pkg/auth/handler"
	forgotpwdhandler "github.com/skygeario/skygear-server/pkg/auth/handler/forgotpwd"
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
	StandaloneTenantConfigurationFile string              `envconfig:"STANDALONE_TENANT_CONFIG_FILE" default:"standalone-tenant-config.yaml"`
	PathPrefix                        string              `envconfig:"PATH_PREFIX"`
	Host                              string              `default:"localhost:3000"`
	ValidHosts                        string              `envconfig:"VALID_HOSTS"`
	Redis                             redis.Configuration `envconfig:"REDIS"`
	InsecureCookie                    bool                `envconfig:"INSECURE_COOKIE"`
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
	envErr := godotenv.Load()
	if envErr != nil {
		log.Print("Error in loading .env file")
	}

	configuration := configuration{}
	envconfig.Process("", &configuration)
	if configuration.ValidHosts == "" {
		configuration.ValidHosts = configuration.Host
	}
	if configuration.Redis.Host == "" {
		log.Fatal("REDIS_HOST is not provided")
	}

	// default template initialization
	templateEngine := template.NewEngine()
	authTemplate.RegisterDefaultTemplates(templateEngine)

	// logging initialization
	logging.SetModule("auth")

	asyncTaskExecutor := async.NewExecutor()
	dbPool := db.NewPool()
	redisPool := redis.NewPool(configuration.Redis)
	authDependency := auth.DependencyMap{
		AsyncTaskExecutor: asyncTaskExecutor,
		TemplateEngine:    templateEngine,
	}

	task.AttachVerifyCodeSendTask(asyncTaskExecutor, authDependency)
	task.AttachPwHousekeeperTask(asyncTaskExecutor, authDependency)
	task.AttachWelcomeEmailSendTask(asyncTaskExecutor, authDependency)

	authContextResolverFactory := resolver.AuthContextResolverFactory{}

	setupCtxFn := server.SetupContextFunc(func(ctx context.Context) context.Context {
		return redis.WithRedis(ctx, redisPool)
	})
	cleanupCtxFn := server.CleanupContextFunc(func(ctx context.Context) {
		redis.CloseConn(ctx)
	})

	var srv server.Server
	if configuration.Standalone {
		filename := configuration.StandaloneTenantConfigurationFile
		reader, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
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
			log.Fatal(err)
		}

		serverOption := server.DefaultOption()
		serverOption.GearPathPrefix = configuration.PathPrefix
		srv = server.NewServerWithOption(configuration.Host, authContextResolverFactory, dbPool, setupCtxFn, cleanupCtxFn, serverOption)
		srv.Use(middleware.TenantConfigurationMiddleware{
			ConfigurationProvider: middleware.ConfigurationProviderFunc(func(_ *http.Request) (config.TenantConfiguration, error) {
				return *tenantConfig, nil
			}),
		}.Handle)
		srv.Use(middleware.ValidateHostMiddleware{ValidHosts: configuration.ValidHosts}.Handle)
		srv.Use(middleware.RequestIDMiddleware{}.Handle)
		srv.Use(middleware.CORSMiddleware{}.Handle)
	} else {
		srv = server.NewServer(configuration.Host, authContextResolverFactory, dbPool, setupCtxFn, cleanupCtxFn)
	}

	handler.AttachSignupHandler(&srv, authDependency)
	handler.AttachLoginHandler(&srv, authDependency)
	handler.AttachLogoutHandler(&srv, authDependency)
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

	go func() {
		log.Printf("Auth gear boot")
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	// wait interrupt signal
	select {
	case <-sig:
		log.Printf("Stoping http server ...\n")
	}

	// create shutdown context with 10 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// shutdown the server
	err := srv.Shutdown(ctx)
	if err != nil {
		log.Printf("Shutdown request error: %v\n", err)
	}
}
