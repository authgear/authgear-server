package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	coreConfig "github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	coreMiddleware "github.com/skygeario/skygear-server/pkg/core/middleware"
	"github.com/skygeario/skygear-server/pkg/core/redis"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/gateway"
	gatewayConfig "github.com/skygeario/skygear-server/pkg/gateway/config"
	"github.com/skygeario/skygear-server/pkg/gateway/handler"
	"github.com/skygeario/skygear-server/pkg/gateway/middleware"
	"github.com/skygeario/skygear-server/pkg/gateway/provider"
	"github.com/skygeario/skygear-server/pkg/gateway/store"
	pqStore "github.com/skygeario/skygear-server/pkg/gateway/store/pq"
	standaloneStore "github.com/skygeario/skygear-server/pkg/gateway/store/standalone"
)

var config gatewayConfig.Configuration

func init() {
	// logging initialization
	logging.SetModule("gateway")

	logger := logging.LoggerEntry("gateway")
	if err := config.ReadFromEnv(); err != nil {
		logger.WithError(err).Panic(
			"Fail to load config for starting gateway server")
	}

	logger.WithField("config", config).Debug("Gateway config")
}

func main() {
	logger := logging.LoggerEntry("gateway")

	// create gateway store
	var store store.GatewayStore
	var connErr error
	if config.Standalone {
		filename := config.StandaloneTenantConfigurationFile
		reader, err := os.Open(filename)
		if err != nil {
			logger.WithError(err).Panic("Fail to open config file")
		}
		tenantConfig, err := coreConfig.NewTenantConfigurationFromYAML(reader)
		if err != nil {
			logger.WithError(err).Panic("Fail to load config from YAML")
		}
		store = &standaloneStore.Store{
			TenantConfig: *tenantConfig,
		}
	} else {
		store, connErr = pqStore.NewGatewayStore(
			context.Background(),
			config.ConnectionStr,
			logger,
		)
		if connErr != nil {
			logger.WithError(connErr).Panic("Fail to create db conn")
		}
	}
	defer store.Close()

	gatewayDependency := gateway.DependencyMap{
		UseInsecureCookie: config.UseInsecureCookie,
	}
	dbPool := db.NewPool()
	redisPool := redis.NewPool(config.Redis)

	rr := mux.NewRouter()
	rr.HandleFunc("/_healthz", HealthCheckHandler)

	r := rr.PathPrefix("/").Subrouter()
	// RecoverMiddleware must come first
	r.Use(coreMiddleware.RecoverMiddleware{
		RecoverHandler: server.DefaultRecoverPanicHandler,
	}.Handle)

	r.Use(coreMiddleware.RequestIDMiddleware{}.Handle)
	r.Use(middleware.FindAppMiddleware{Store: store}.Handle)

	gr := r.PathPrefix("/_{gear}").Subrouter()

	gr.Use(coreMiddleware.TenantConfigurationMiddleware{
		ConfigurationProvider: provider.GatewayTenantConfigurationProvider{
			Store: store,
		},
	}.Handle)
	gr.Use(middleware.TenantAuthzMiddleware{
		Store:         store,
		Configuration: config,
	}.Handle)
	gr.Use(coreMiddleware.CORSMiddleware{}.Handle)

	gr.HandleFunc("/{rest:.*}", handler.NewGearHandler("rest"))

	cr := r.PathPrefix("/").Subrouter()

	cr.Use(coreMiddleware.TenantConfigurationMiddleware{
		ConfigurationProvider: provider.GatewayTenantConfigurationProvider{
			Store: store,
		},
	}.Handle)

	cr.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = db.InitRequestDBContext(r, dbPool)
			r = auth.InitRequestAuthContext(r)
			r = r.WithContext(redis.WithRedis(r.Context(), redisPool))
			next.ServeHTTP(w, r)
		})
	})

	cr.Use(middleware.FindDeploymentRouteMiddleware{
		RestPathIdentifier: "rest",
		Store:              store,
	}.Handle)

	cr.Use(coreMiddleware.Injecter{
		MiddlewareFactory: coreMiddleware.AuthnMiddlewareFactory{},
		Dependency:        gatewayDependency,
	}.Handle)

	cr.Use(coreMiddleware.Injecter{
		MiddlewareFactory: middleware.AuthInfoMiddlewareFactory{},
		Dependency:        gatewayDependency,
	}.Handle)

	cr.Use(coreMiddleware.CORSMiddleware{}.Handle)

	cr.HandleFunc("/{rest:.*}", handler.NewDeploymentRouteHandler())

	srv := &http.Server{
		Addr: config.Host,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      rr, // Pass our instance of gorilla/mux in.
	}

	logger.Info("Start gateway server")
	if err := srv.ListenAndServe(); err != nil {
		logger.Errorf("Fail to start gateway server %v", err)
	}
}

// HealthCheckHandler is basic handler for server health check
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, "OK")
}
