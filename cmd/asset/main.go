package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"

	"github.com/skygeario/skygear-server/pkg/asset"
	"github.com/skygeario/skygear-server/pkg/asset/config"
	"github.com/skygeario/skygear-server/pkg/core/cloudstorage"
	coreConfig "github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/middleware"
	"github.com/skygeario/skygear-server/pkg/core/redis"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

/*
	@API Asset Gear
	@Version 1.0.0
	@Server {base_url}/_asset
		Asset Gear URL
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
*/
func main() {
	logging.SetModule("asset")
	loggerFactory := logging.NewFactory(logging.NewDefaultMaskedTextFormatter(nil))
	logger := loggerFactory.NewLogger("asset")

	if err := godotenv.Load(); err != nil {
		logger.WithError(err).Debug("Cannot load .env file")
	}

	configuration := config.Configuration{}
	envconfig.Process("", &configuration)
	if err := configuration.Initialize(); err != nil {
		logger.WithError(err).Panic("cannot initialize configuration")
	}

	var storage cloudstorage.Storage
	switch configuration.Storage.Backend {
	case config.StorageBackendAzure:
		storage = cloudstorage.NewAzureStorage(
			configuration.Storage.Azure.StorageAccount,
			configuration.Storage.Azure.AccessKey,
			configuration.Storage.Azure.Container,
		)
	case config.StorageBackendGCS:
		storage = cloudstorage.NewGCSStorage(
			configuration.Storage.GCS.ServiceAccount,
			configuration.Storage.GCS.PrivateKey,
			configuration.Storage.GCS.Bucket,
		)
	case config.StorageBackendS3:
		storage = cloudstorage.NewS3Storage(
			configuration.Storage.S3.AccessKey,
			configuration.Storage.S3.SecretKey,
			configuration.Storage.S3.Region,
			configuration.Storage.S3.Bucket,
		)
	}

	dbPool := db.NewPool()
	redisPool, err := redis.NewPool(configuration.Redis)
	if err != nil {
		logger.Fatalf("fail to create redis pool: %v", err)
	}
	dependencyMap := &asset.DependencyMap{
		UseInsecureCookie: configuration.UseInsecureCookie,
		Storage:           storage,
	}

	var srv server.Server
	if configuration.Standalone {
		filename := configuration.StandaloneTenantConfigurationFile
		reader, err := os.Open(filename)
		if err != nil {
			logger.WithError(err).Error("Cannot open standalone config")
		}
		tenantConfig, err := coreConfig.NewTenantConfigurationFromYAML(reader)
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
		serverOption.GearPathPrefix = "/_asset"
		srv = server.NewServerWithOption(configuration.ServerHost, dependencyMap, serverOption)
		srv.Use(middleware.TenantConfigurationMiddleware{
			ConfigurationProvider: middleware.ConfigurationProviderFunc(func(_ *http.Request) (coreConfig.TenantConfiguration, error) {
				return *tenantConfig, nil
			}),
		}.Handle)
		srv.Use(middleware.RequestIDMiddleware{}.Handle)
		srv.Use(middleware.CORSMiddleware{}.Handle)
	} else {
		srv = server.NewServer(configuration.ServerHost, dependencyMap)
	}

	srv.Use(middleware.DBMiddleware{Pool: dbPool}.Handle)
	srv.Use(middleware.RedisMiddleware{Pool: redisPool}.Handle)
	srv.Use(middleware.AuthMiddleware{}.Handle)
	srv.Use(middleware.Injecter{
		MiddlewareFactory: middleware.AuthnMiddlewareFactory{},
		Dependency:        dependencyMap,
	}.Handle)

	go func() {
		logger.Info("Starting asset gear")
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
	err = srv.Shutdown(ctx)
	if err != nil {
		logger.WithError(err).Fatal("Cannot shutdown server")
	}
}
