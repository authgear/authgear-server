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

	"github.com/authgear/authgear-server/pkg/auth"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
	"github.com/authgear/authgear-server/pkg/resolver"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/version"
	"github.com/authgear/authgear-server/pkg/worker"
)

type Controller struct {
	ServeMain     bool
	ServeResolver bool

	logger   *log.Logger
	ctx      context.Context
	shutdown <-chan struct{}
	wg       *sync.WaitGroup
}

type serverType string

const (
	serverMain     serverType = "Main Server"
	serverResolver serverType = "Resolver Server"
)

func (c *Controller) Start() {
	cfg, err := config.LoadServerConfigFromEnv()
	if err != nil {
		golog.Fatalf("failed to load server config: %s", err)
	}

	var wrk *worker.Worker
	taskQueueFactory := deps.TaskQueueFactory(func(provider *deps.AppProvider) task.Queue {
		return newInProcessQueue(provider, wrk.Executor)
	})

	p, err := deps.NewRootProvider(cfg, taskQueueFactory)
	if err != nil {
		golog.Fatalf("failed to setup server: %s", err)
	}

	wrk = worker.NewWorker(p)

	// From now, we should use c.logger to log.
	c.logger = p.LoggerFactory.New("server")

	c.logger.Infof("authgear (version %s)", version.Version)
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

	if c.ServeMain {
		c.startServer(cfg, serverMain, cfg.ListenAddr, auth.NewRouter(p, configSource))
	}
	if c.ServeResolver {
		c.startServer(cfg, serverResolver, cfg.ResolverListenAddr, resolver.NewRouter(p, configSource))
	}

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

func (c *Controller) startServer(cfg *config.ServerConfig, t serverType, addr string, handler http.Handler) {
	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	go func() {
		c.logger.Infof("starting %s on %v", t, addr)
		var err error
		if cfg.DevMode && t == serverMain {
			err = server.ListenAndServeTLS(cfg.TLSCertFilePath, cfg.TLSKeyFilePath)
		} else {
			err = server.ListenAndServe()
		}

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			c.logger.WithError(err).Fatalf("failed to start %s", t)
		}
	}()

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		select {
		case <-c.shutdown:
			break
		}
		c.logger.Infof("stopping %s...", t)

		err := server.Shutdown(c.ctx)
		if err != nil {
			c.logger.WithError(err).Errorf("failed to stop %s gracefully", t)
		}
	}()
}
