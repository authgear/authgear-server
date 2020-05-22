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

	coreConfig "github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	coreMiddleware "github.com/skygeario/skygear-server/pkg/core/middleware"
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
	dbPool := db.NewPool()

	// create gateway store
	var store store.GatewayStore
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
			TenantConfig:  *tenantConfig,
			GatewayConfig: config,
		}
	} else {
		var err error
		store, err = pqStore.NewGatewayStore(
			context.Background(),
			dbPool,
			config.ConnectionStr,
		)
		if err != nil {
			logger.WithError(err).Panic("Fail to create gateway store")
		}
	}
	defer store.Close()

	gatewayDependency := gateway.DependencyMap{
		Config: config,
	}
	rr := mux.NewRouter()
	rr.HandleFunc("/healthz", HealthCheckHandler)

	r := rr.PathPrefix("/").Subrouter()
	r.Use(sentry.Middleware(sentry.DefaultClient.Hub))
	r.Use(coreMiddleware.RecoverMiddleware{}.Handle)

	r.Use(middleware.FindAppMiddleware{Store: store}.Handle)

	r.Use(coreMiddleware.WriteTenantConfigMiddleware{
		ConfigurationProvider: provider.GatewayTenantConfigurationProvider{
			Store: store,
		},
	}.Handle)

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = db.InitRequestDBContext(r, dbPool)
			next.ServeHTTP(w, r)
		})
	})

	// CORS headers should be set right after a proxy backend has been found.
	r.Use(coreMiddleware.CORSMiddleware{}.Handle)

	r.Use(coreMiddleware.Injecter{
		MiddlewareFactory: middleware.AuthnMiddlewareFactory{},
		Dependency:        gatewayDependency,
	}.Handle)

	r.Handle("/{rest:.*}", handler.NewGatewayHandler(gatewayDependency))

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
