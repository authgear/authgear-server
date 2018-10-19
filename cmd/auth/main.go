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

	authDependency := auth.NewDependencyMap()

	authContextResolverFactory := resolver.AuthContextResolverFactory{}
	srv := server.NewServer(configuration.Host, authContextResolverFactory)

	if configuration.DevMode {
		srv.Use(middleware.TenantConfigurationMiddleware{
			ConfigurationProvider: middleware.ConfigurationProviderFunc(config.NewTenantConfigurationFromEnv),
		}.Handle)
	}

	srv.Use(middleware.RequestIDMiddleware{}.Handle)
	srv.Use(middleware.CORSMiddleware{}.Handle)
	handler.AttachSignupHandler(&srv, authDependency)
	handler.AttachLoginHandler(&srv, authDependency)
	handler.AttachLogoutHandler(&srv, authDependency)
	handler.AttachMeHandler(&srv, authDependency)
	handler.AttachSetDisableHandler(&srv, authDependency)
	handler.AttachRoleAssignHandler(&srv, authDependency)
	handler.AttachRoleRevokeHandler(&srv, authDependency)
	handler.AttachResetPasswordHandler(&srv, authDependency)
	handler.AttachGetRoleHandler(&srv, authDependency)
	handler.AttachRoleAdminHandler(&srv, authDependency)
	handler.AttachRoleDefaultHandler(&srv, authDependency)

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
