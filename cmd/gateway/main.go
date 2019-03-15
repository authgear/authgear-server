package main

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/core/logging"
	coreMiddleware "github.com/skygeario/skygear-server/pkg/core/middleware"
	"github.com/skygeario/skygear-server/pkg/core/server"
	gatewayConfig "github.com/skygeario/skygear-server/pkg/gateway/config"
	pqStore "github.com/skygeario/skygear-server/pkg/gateway/db/pq"
	"github.com/skygeario/skygear-server/pkg/gateway/handler"
	"github.com/skygeario/skygear-server/pkg/gateway/middleware"
	"github.com/skygeario/skygear-server/pkg/gateway/provider"
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
	store, connErr := pqStore.NewGatewayStore(
		context.Background(),
		config.DB.ConnectionStr,
	)
	if connErr != nil {
		logger.WithError(connErr).Panic("Fail to create db conn")
	}
	defer store.Close()

	r := mux.NewRouter()
	r.HandleFunc("/healthz", HealthCheckHandler)

	r = r.PathPrefix("/").Subrouter()
	// RecoverMiddleware must come first
	r.Use(coreMiddleware.RecoverMiddleware{
		RecoverHandler: server.DefaultRecoverPanicHandler,
	}.Handle)

	r.Use(middleware.FindAppMiddleware{Store: store}.Handle)

	gr := r.PathPrefix("/_{gear}").Subrouter()

	gr.Use(coreMiddleware.TenantConfigurationMiddleware{
		ConfigurationProvider: provider.GatewayTenantConfigurationProvider{
			Store: store,
		},
	}.Handle)
	gr.Use(middleware.TenantAuthzMiddleware{
		Store:        store,
		RouterConfig: config.Router,
	}.Handle)

	gr.HandleFunc("/{rest:.*}", handler.NewGearHandler("rest"))

	cr := r.PathPrefix("/").Subrouter()

	cr.Use(middleware.FindCloudCodeMiddleware{
		RestPathIdentifier: "rest",
		Store:              store,
	}.Handle)

	cr.HandleFunc("/{rest:.*}", handler.NewCloudCodeHandler(config.Router))

	srv := &http.Server{
		Addr: config.HTTP.Host,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r, // Pass our instance of gorilla/mux in.
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
