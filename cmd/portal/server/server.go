package server

import (
	"context"
	golog "log"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
	"github.com/authgear/authgear-server/pkg/lib/infra/task/executor"
	"github.com/authgear/authgear-server/pkg/lib/infra/task/queue"
	"github.com/authgear/authgear-server/pkg/portal"
	"github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/server"
	"github.com/authgear/authgear-server/pkg/version"
	"github.com/authgear/authgear-server/pkg/worker"
)

type Controller struct {
	logger *log.Logger
}

func ProvideCaptureTaskContext(config *config.Config, appCtx *config.AppContext) task.CaptureTaskContext {
	return func() *task.Context {
		return &task.Context{
			Config:     config,
			AppContext: appCtx,
		}
	}
}

func newInProcessQueue(p *deps.AppProvider, e *executor.InProcessExecutor) *queue.InProcessQueue {
	handle := p.AppDatabase
	appContext := p.AppContext
	config := appContext.Config
	captureTaskContext := ProvideCaptureTaskContext(config, appContext)
	inProcessQueue := &queue.InProcessQueue{
		Database:       handle,
		CaptureContext: captureTaskContext,
		Executor:       e,
	}
	return inProcessQueue
}

func (c *Controller) Start() {
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		golog.Fatalf("failed to load server config: %s", err)
	}

	var wrk *worker.Worker
	taskQueueFactory := deps.TaskQueueFactory(func(provider *deps.AppProvider) task.Queue {
		return newInProcessQueue(provider, wrk.Executor)
	})

	p, err := deps.NewRootProvider(
		cfg.EnvironmentConfig,
		cfg.BuiltinResourceDirectory,
		cfg.CustomResourceDirectory,
		cfg.App.BuiltinResourceDirectory,
		cfg.App.CustomResourceDirectory,
		cfg.ConfigSource,
		&cfg.Authgear,
		&cfg.AdminAPI,
		&cfg.App,
		&cfg.SMTP,
		&cfg.Mail,
		&cfg.Kubernetes,
		cfg.DomainImplementation,
		&cfg.Search,
		&cfg.Web3,
		&cfg.AuditLog,
		&cfg.Analytic,
		&cfg.Stripe,
		&cfg.GoogleTagManager,
		taskQueueFactory,
	)
	if err != nil {
		golog.Fatalf("failed to setup server: %s", err)
	}

	// From now, we should use c.logger to log.
	c.logger = p.LoggerFactory.New("authgear-portal")

	c.logger.Infof("authgear-portal (version %s)", version.Version)
	if cfg.DevMode {
		c.logger.Warn("development mode is ON - do not use in production")
	}

	configSrcController := newConfigSourceController(p, context.Background())
	err = configSrcController.Open()
	if err != nil {
		c.logger.WithError(err).Fatal("cannot open configuration")
	}
	defer configSrcController.Close()

	p.ConfigSourceController = configSrcController

	var specs []server.Spec
	specs = append(specs, server.Spec{
		Name:          "portal server",
		ListenAddress: cfg.PortalListenAddr,
		Handler:       portal.NewRouter(p),
	})
	server.Start(c.logger, specs)
}
