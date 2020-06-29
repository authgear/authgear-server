package server

import (
	"context"
	"errors"
	golog "log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/config"
	"github.com/skygeario/skygear-server/pkg/deps"
	"github.com/skygeario/skygear-server/pkg/log"
)

type Controller struct {
	logger   *log.Logger
	ctx      context.Context
	shutdown <-chan struct{}
	wg       *sync.WaitGroup
}

func (c *Controller) Start() {
	cfg, err := config.LoadServerConfigFromEnv()
	if err != nil {
		golog.Fatalf("failed to load server config: %s", err)
	}

	p, err := deps.NewRootProvider(cfg)
	if err != nil {
		golog.Fatalf("failed to setup server: %s", err)
	}

	// From now, we should use c.logger to log.
	c.logger = p.LoggerFactory.New("server")

	if cfg.DevMode {
		c.logger.Warn("development mode is ON - do not use in production")
	}

	configSource := newConfigSource(p)
	err = configSource.Open()
	if err != nil {
		c.logger.WithError(err).Fatal("cannot open configuration")
	}

	c.wg = new(sync.WaitGroup)
	shutdown := make(chan struct{})
	c.shutdown = shutdown

	setupTasks(p.TaskExecutor, p)
	c.startServer("public server", cfg.PublicListenAddr, setupRoutes(p, configSource))
	c.startServer("internal server", cfg.InternalListenAddr, setupInternalRoutes(p, configSource))

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sig:
		c.logger.Infof("received signal %s, shutting down...", sig.String())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	c.ctx = ctx
	close(shutdown)
	c.wg.Wait()
}

func (c *Controller) startServer(name string, addr string, handler http.Handler) {
	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	go func() {
		c.logger.Infof("starting %s on %v", name, addr)
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			c.logger.WithError(err).Fatalf("failed to start %s", name)
		}
	}()

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		select {
		case <-c.shutdown:
			break
		}
		c.logger.Infof("stopping %s...", name)

		err := server.Shutdown(c.ctx)
		if err != nil {
			c.logger.WithError(err).Errorf("failed to stop %s gracefully", name)
		}
	}()
}
