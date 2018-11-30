package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authn/resolver"

	"github.com/kelseyhightower/envconfig"

	"github.com/joho/godotenv"
	"github.com/skygeario/skygear-server/pkg/auth/handler"
	"github.com/skygeario/skygear-server/pkg/auth/handler/ssohandler"
	"github.com/skygeario/skygear-server/pkg/auth/task"
	asyncServer "github.com/skygeario/skygear-server/pkg/core/async/server"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/middleware"
	"github.com/skygeario/skygear-server/pkg/core/server"
)

type configuration struct {
	DevMode bool   `envconfig:"DEV_MODE"`
	Host    string `envconfig:"HOST" default:"localhost:3000"`
}

func main() {
	envErr := godotenv.Load()
	if envErr != nil {
		log.Print("Error in loading .env file")
	}

	configuration := configuration{}
	envconfig.Process("", &configuration)

	// logging initialization
	logging.SetModule("auth")

	authDependency := auth.DependencyMap{}

	asyncTaskServer := asyncServer.NewTaskServer()
	task.AttachVerifyCodeSendTask(asyncTaskServer, authDependency)

	authRequestDependency := auth.RequestDependencyMap{
		DependencyMap:   authDependency,
		AsyncTaskServer: asyncTaskServer,
	}

	authContextResolverFactory := resolver.AuthContextResolverFactory{}
	srv := server.NewServer(configuration.Host, authContextResolverFactory)

	if configuration.DevMode {
		srv.Use(middleware.TenantConfigurationMiddleware{
			ConfigurationProvider: middleware.ConfigurationProviderFunc(config.NewTenantConfigurationFromEnv),
		}.Handle)
	}

	srv.Use(middleware.RequestIDMiddleware{}.Handle)
	srv.Use(middleware.CORSMiddleware{}.Handle)

	handler.AttachSignupHandler(&srv, authRequestDependency)
	handler.AttachLoginHandler(&srv, authRequestDependency)
	handler.AttachLogoutHandler(&srv, authRequestDependency)
	handler.AttachMeHandler(&srv, authRequestDependency)
	handler.AttachSetDisableHandler(&srv, authRequestDependency)
	handler.AttachRoleAssignHandler(&srv, authRequestDependency)
	handler.AttachRoleRevokeHandler(&srv, authRequestDependency)
	handler.AttachResetPasswordHandler(&srv, authRequestDependency)
	handler.AttachGetRoleHandler(&srv, authRequestDependency)
	handler.AttachRoleAdminHandler(&srv, authRequestDependency)
	handler.AttachRoleDefaultHandler(&srv, authRequestDependency)
	handler.AttachWelcomeEmailHandler(&srv, authRequestDependency)
	handler.AttachForgotPasswordHandler(&srv, authRequestDependency)
	handler.AttachVerifyRequestHandler(&srv, authRequestDependency)
	ssohandler.AttachAuthURLHandler(&srv, authRequestDependency)
	ssohandler.AttachConfigHandler(&srv, authRequestDependency)
	ssohandler.AttachIFrameHandlerFactory(&srv, authRequestDependency)
	ssohandler.AttachCustomTokenLoginHandler(&srv, authRequestDependency)

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
