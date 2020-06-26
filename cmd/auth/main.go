package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/skygeario/skygear-server/pkg/auth/config"
	"github.com/skygeario/skygear-server/pkg/deps"
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

	publicServer := &http.Server{
		Addr:    serverCfg.PublicListenAddr,
		Handler: setupRoutes(p, configSource),
	}

	internalServer := &http.Server{
		Addr:    serverCfg.InternalListenAddr,
		Handler: setupInternalRoutes(p, configSource),
	}

	go func() {
		logger.Infof("starting public server on %v", serverCfg.PublicListenAddr)
		err := publicServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.WithError(err).Fatal("failed to start public server")
		}
	}()

	go func() {
		logger.Infof("starting internal server on %v", serverCfg.InternalListenAddr)
		err := internalServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.WithError(err).Fatal("failed to start internal server")
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sig:
		logger.Info("stopping gracefully")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var wg sync.WaitGroup

	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		err := publicServer.Shutdown(ctx)
		if err != nil {
			logger.WithError(err).Error("failed to stop public server gracefully")
		}
	}(&wg)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		err := internalServer.Shutdown(ctx)
		if err != nil {
			logger.WithError(err).Error("failed to stop internal server gracefully")
		}
	}(&wg)

	wg.Wait()
}
