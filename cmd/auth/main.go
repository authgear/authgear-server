package main

import (
	"github.com/skygeario/skygear-server/pkg/auth/config"
	"github.com/skygeario/skygear-server/pkg/deps"
	"github.com/skygeario/skygear-server/pkg/httputil"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("failed to load .env file: %s", err)
	}

	serverCfg, err := config.LoadServerConfigFromEnv()
	if err != nil {
		log.Fatalf("failed to load server config: %s", err)
	}

	rootDeps, err := deps.NewRootContainer(serverCfg)
	if err != nil {
		log.Fatalf("failed to setup server: %s", err)
	}

	logger := rootDeps.LoggerFactory.New("main")

	if serverCfg.DevMode {
		logger.Warn("Development mode is ON - do not use in production")
	}

	server := httputil.NewServer(rootDeps.LoggerFactory, &http.Server{
		Addr:    serverCfg.ListenAddr,
		Handler: nil,
	})
	server.ListenAndServe("starting auth gear")
}
