package server

import (
	"context"
	golog "log"

	"github.com/authgear/authgear-server/pkg/portal"
	"github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/pprofutil"
	"github.com/authgear/authgear-server/pkg/util/server"
	"github.com/authgear/authgear-server/pkg/util/signalutil"
	"github.com/authgear/authgear-server/pkg/version"
)

type Controller struct {
	logger *log.Logger
}

func (c *Controller) Start(ctx context.Context) {
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		golog.Fatalf("failed to load server config: %s", err)
	}

	p, err := deps.NewRootProvider(
		cfg.EnvironmentConfig,
		cfg.CustomResourceDirectory,
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
		&cfg.AuditLog,
		&cfg.Analytic,
		&cfg.Stripe,
		&cfg.Osano,
		&cfg.GoogleTagManager,
		&cfg.PortalFrontendSentry,
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

	configSrcController := newConfigSourceController(p)
	err = configSrcController.Open(ctx)
	if err != nil {
		c.logger.WithError(err).Fatal("cannot open configuration")
	}
	defer configSrcController.Close()

	p.ConfigSourceController = configSrcController

	var specs []signalutil.Daemon
	specs = append(specs, server.NewSpec(ctx, &server.Spec{
		Name:          "authgear-portal",
		ListenAddress: cfg.PortalListenAddr,
		Handler:       portal.NewRouter(p),
	}))
	specs = append(specs, server.NewSpec(ctx, &server.Spec{
		Name:          "authgear-portal-internal",
		ListenAddress: cfg.PortalInternalListenAddr,
		Handler:       pprofutil.NewServeMux(),
	}))
	signalutil.Start(ctx, c.logger, specs...)
}
