package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"

	"github.com/skygeario/skygear-server/pkg/core/auth/authn/resolver"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/middleware"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/record"
	"github.com/skygeario/skygear-server/pkg/record/handler"
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
	logging.SetModule("record")

	recordDependency := record.NewDependencyMap()

	authContextResolverFactory := resolver.AuthContextResolverFactory{}
	srv := server.NewServer(configuration.Host, authContextResolverFactory)

	if configuration.DevMode {
		srv.Use(middleware.TenantConfigurationMiddleware{
			ConfigurationProvider: middleware.ConfigurationProviderFunc(config.NewTenantConfigurationFromEnv),
		}.Handle)
	}

	srv.Use(middleware.RequestIDMiddleware{}.Handle)

	handler.AttachSaveHandler(&srv, recordDependency)
	handler.AttachFetchHandler(&srv, recordDependency)
	handler.AttachQueryHandler(&srv, recordDependency)

	handler.AttachSchemaCreateHandler(&srv, recordDependency)
	handler.AttachSchemaDeleteHandler(&srv, recordDependency)
	handler.AttachSchemaRenameHandler(&srv, recordDependency)

	go func() {
		log.Printf("Record gear boot")
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
