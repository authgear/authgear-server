package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	coreConfig "github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	coreMiddleware "github.com/skygeario/skygear-server/pkg/core/middleware"
	"github.com/skygeario/skygear-server/pkg/core/redis"
	"github.com/skygeario/skygear-server/pkg/core/sentry"
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
var loggerFactory logging.Factory
var logger *logrus.Entry

func init() {
	// logging initialization
	logging.SetModule("gateway")
	loggerFactory = logging.NewFactory(
		logging.NewDefaultLogHook(nil),
		&sentry.LogHook{Hub: sentry.DefaultClient.Hub},
	)
	logger = loggerFactory.NewLogger("gateway")

	if err := godotenv.Load(); err != nil {
		logger.WithError(err).Info(
			"Error in loading .env file, continue without .env")
	}

	if err := config.ReadFromEnv(); err != nil {
		logger.WithError(err).Panic(
			"Fail to load config for starting gateway server")
	}
}

func main() {
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
			loggerFactory,
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
	redisPool, err := redis.NewPool(config.Redis)
	if err != nil {
		logger.Fatalf("fail to create redis pool: %v", err.Error())
	}
	rr := mux.NewRouter()
	rr.HandleFunc("/_healthz", HealthCheckHandler)

	r := rr.PathPrefix("/").Subrouter()
	r.Use(sentry.Middleware(sentry.DefaultClient.Hub))
	r.Use(coreMiddleware.RecoverMiddleware{}.Handle)

	r.Use(coreMiddleware.RequestIDMiddleware{}.Handle)
	r.Use(middleware.FindAppMiddleware{Store: store}.Handle)

	gr := r.PathPrefix("/_{gear}").Subrouter()

	gr.Use(coreMiddleware.WriteTenantConfigMiddleware{
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

	cr.Use(coreMiddleware.WriteTenantConfigMiddleware{
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

	// CORS headers should be set right after a proxy backend has been found.
	cr.Use(coreMiddleware.CORSMiddleware{}.Handle)

	cr.Use(coreMiddleware.Injecter{
		MiddlewareFactory: coreMiddleware.AuthnMiddlewareFactory{},
		Dependency:        gatewayDependency,
	}.Handle)

	cr.Use(coreMiddleware.Injecter{
		MiddlewareFactory: middleware.AuthInfoMiddlewareFactory{},
		Dependency:        gatewayDependency,
	}.Handle)

	cr.HandleFunc("/{rest:.*}", handler.NewDeploymentRouteHandler())

	srv := &http.Server{
		Addr:         config.Host,
		ReadTimeout:  time.Second * 60,
		WriteTimeout: time.Second * 60,
		IdleTimeout:  time.Second * 60,
		Handler:      rr, // Pass our instance of gorilla/mux in.
	}

	logger.Info("Start gateway server")
	if err := srv.ListenAndServe(); err != nil {
		logger.WithError(err).Errorf("Fail to start gateway server")
	}
}

// HealthCheckHandler is basic handler for server health check
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, "OK")
}
