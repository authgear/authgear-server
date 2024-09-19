package server

import (
	"context"
	golog "log"

	"github.com/authgear/authgear-server/pkg/admin"
	"github.com/authgear/authgear-server/pkg/auth"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	infraredisqueue "github.com/authgear/authgear-server/pkg/lib/infra/redisqueue"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
	"github.com/authgear/authgear-server/pkg/redisqueue"
	"github.com/authgear/authgear-server/pkg/resolver"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/pprofutil"
	"github.com/authgear/authgear-server/pkg/util/server"
	"github.com/authgear/authgear-server/pkg/util/signalutil"
	"github.com/authgear/authgear-server/pkg/version"
	"github.com/authgear/authgear-server/pkg/worker"
)

type Controller struct {
	ServeMain     bool
	ServeResolver bool
	ServeAdmin    bool

	logger *log.Logger
}

func (c *Controller) Start() {
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		golog.Fatalf("failed to load server config: %v", err)
	}

	var wrk *worker.Worker
	taskQueueFactory := deps.TaskQueueFactory(func(provider *deps.AppProvider) task.Queue {
		return newInProcessQueue(provider, wrk.Executor)
	})

	p, err := deps.NewRootProvider(
		cfg.EnvironmentConfig,
		cfg.ConfigSource,
		cfg.BuiltinResourceDirectory,
		cfg.CustomResourceDirectory,
		taskQueueFactory,
	)
	if err != nil {
		golog.Fatalf("failed to setup server: %v", err)
	}

	wrk = worker.NewWorker(p)

	// From now, we should use c.logger to log.
	c.logger = p.LoggerFactory.New("server")

	c.logger.Infof("authgear (version %s)", version.Version)
	if cfg.DevMode {
		c.logger.Warn("development mode is ON - do not use in production")
	}

	configSrcController := newConfigSourceController(p, context.Background())
	err = configSrcController.Open()
	if err != nil {
		c.logger.WithError(err).Fatal("cannot open configuration")
	}
	defer configSrcController.Close()

	var specs []signalutil.Daemon

	if c.ServeMain {
		u, err := server.ParseListenAddress(cfg.MainListenAddr)
		if err != nil {
			c.logger.WithError(err).Fatal("failed to parse main server listen address")
		}

		spec := &server.Spec{
			Name:          "Main Server",
			ListenAddress: u.Host,
			Handler: auth.NewRouter(
				p,
				configSrcController.GetConfigSource(),
			),
		}

		if cfg.DevMode && u.Scheme == "https" {
			spec.HTTPS = true
			spec.CertFilePath = cfg.TLSCertFilePath
			spec.KeyFilePath = cfg.TLSKeyFilePath
		}

		specs = append(specs, server.NewSpec(spec))

		// Set up internal server.
		u, err = server.ParseListenAddress(cfg.MainInteralListenAddr)
		if err != nil {
			c.logger.WithError(err).Fatal("failed to parse main server internal listen address")
		}
		specs = append(specs, server.NewSpec(&server.Spec{
			Name:          "Main Internal Server",
			ListenAddress: u.Host,
			Handler:       pprofutil.NewServeMux(),
		}))

		specs = append(specs, redisqueue.NewConsumer(
			infraredisqueue.QueueUserReindex,
			p,
			configSrcController,
			redisqueue.UserReindex,
		))
	}

	if c.ServeResolver {
		u, err := server.ParseListenAddress(cfg.ResolverListenAddr)
		if err != nil {
			c.logger.WithError(err).Fatal("failed to parse resolver server listen address")
		}

		specs = append(specs, server.NewSpec(&server.Spec{
			Name:          "Resolver Server",
			ListenAddress: u.Host,
			Handler:       resolver.NewRouter(p, configSrcController.GetConfigSource()),
		}))

		// Set up internal server.
		u, err = server.ParseListenAddress(cfg.ResolverInternalListenAddr)
		if err != nil {
			c.logger.WithError(err).Fatal("failed to parse resolver internal server listen address")
		}

		specs = append(specs, server.NewSpec(&server.Spec{
			Name:          "Resolver Internal Server",
			ListenAddress: u.Host,
			Handler:       pprofutil.NewServeMux(),
		}))
	}

	if c.ServeAdmin {
		u, err := server.ParseListenAddress(cfg.AdminListenAddr)
		if err != nil {
			c.logger.WithError(err).Fatal("failed to parse admin API server listen address")
		}

		specs = append(specs, server.NewSpec(&server.Spec{
			Name:          "Admin API Server",
			ListenAddress: u.Host,
			Handler: admin.NewRouter(
				p,
				configSrcController.GetConfigSource(),
				cfg.AdminAPIAuth,
			),
		}))

		u, err = server.ParseListenAddress(cfg.AdminInternalListenAddr)
		if err != nil {
			c.logger.WithError(err).Fatal("failed to parse admin API internal server listen address")
		}

		specs = append(specs, server.NewSpec(&server.Spec{
			Name:          "Admin API Internal Server",
			ListenAddress: u.Host,
			Handler:       pprofutil.NewServeMux(),
		}))

		specs = append(specs, redisqueue.NewConsumer(
			infraredisqueue.QueueUserImport,
			p,
			configSrcController,
			redisqueue.UserImport,
		))

		specs = append(specs, redisqueue.NewConsumer(
			infraredisqueue.QueueUserExport,
			p,
			configSrcController,
			redisqueue.UserExport,
		))
	}

	signalutil.Start(c.logger, specs...)
}
