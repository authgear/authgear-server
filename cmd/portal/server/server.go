package server

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/authgear/authgear-server/pkg/portal"
	"github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/authgear/authgear-server/pkg/util/pprofutil"
	"github.com/authgear/authgear-server/pkg/util/server"
	"github.com/authgear/authgear-server/pkg/util/signalutil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
	"github.com/authgear/authgear-server/pkg/version"
)

var logger = slogutil.NewLogger("authgear-portal")

type Controller struct{}

func (c *Controller) Start(ctx context.Context) {
	logger := logger.GetLogger(ctx)

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		err = fmt.Errorf("failed to load server config: %w", err)
		panic(err)
	}

	ctx, p, err := deps.NewRootProvider(
		ctx,
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
		&cfg.PortalFeatures,
	)

	if err != nil {
		err = fmt.Errorf("failed to setup server: %w", err)
		panic(err)
	}

	logger.Info(ctx, "authgear-portal version", slog.String("version", version.Version))
	if cfg.DevMode {
		logger.Warn(ctx, "development mode is ON - do not use in production")
	}

	configSrcController := newConfigSourceController(p)
	err = configSrcController.Open(ctx)
	if err != nil {
		err = fmt.Errorf("cannot open configuration: %w", err)
		panic(err)
	}
	defer func() {
		_ = configSrcController.Close()
	}()

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
	signalutil.Start(ctx, specs...)
}
