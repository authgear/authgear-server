package server

import (
	golog "log"

	"github.com/authgear/authgear-server/pkg/admin"
	"github.com/authgear/authgear-server/pkg/auth"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
	"github.com/authgear/authgear-server/pkg/resolver"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/server"
	"github.com/authgear/authgear-server/pkg/version"
	"github.com/authgear/authgear-server/pkg/worker"
)

type Controller struct {
	ServeMain     bool
	ServeResolver bool
	ServeAdmin    bool

	logger *log.Logger
}

type serverType string

const (
	serverMain     serverType = "Main Server"
	serverResolver serverType = "Resolver Server"
	serverAdminAPI serverType = "Admin API Server"
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

	var specs []server.Spec

	if c.ServeMain {
		specs = append(specs, server.Spec{
			Name:          string(serverMain),
			ListenAddress: cfg.ListenAddr,
			Handler:       auth.NewRouter(p, configSource),
		})
	}

	if c.ServeResolver {
		specs = append(specs, server.Spec{
			Name:          string(serverResolver),
			ListenAddress: cfg.ResolverListenAddr,
			Handler:       resolver.NewRouter(p, configSource),
		})
	}

	if c.ServeAdmin {
		specs = append(specs, server.Spec{
			Name:          string(serverAdminAPI),
			ListenAddress: cfg.AdminListenAddr,
			Handler:       admin.NewRouter(p, configSource),
		})
	}

	server.Start(c.logger, specs)
}
