package main

import (
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/skygeario/skygear-server/pkg/auth/config"
	"github.com/skygeario/skygear-server/pkg/deps"
	"github.com/skygeario/skygear-server/pkg/httputil"
)

func main() {
	err := godotenv.Load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("failed to load .env file: %s", err)
	}

	serverCfg, err := config.LoadServerConfigFromEnv()
	if err != nil {
		log.Fatalf("failed to load server config: %s", err)
	}

	p, err := deps.NewRootProvider(serverCfg)
	if err != nil {
		log.Fatalf("failed to setup server: %s", err)
	}

	logger := p.LoggerFactory.New("main")

	if serverCfg.DevMode {
		logger.Warn("Development mode is ON - do not use in production")
	}

	configSource := newConfigSource(p)
	err = configSource.Open()
	if err != nil {
		logger.WithError(err).Fatal("cannot open configuration")
	}

	setupTasks(p.TaskExecutor, p)

	server := httputil.NewServer(p.LoggerFactory, &http.Server{
		Addr:    serverCfg.ListenAddr,
		Handler: setupRoutes(p, configSource),
	})
	server.ListenAndServe("starting auth gear")
}
