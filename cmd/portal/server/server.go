package server

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/authgear/authgear-server/pkg/portal"
	"github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/authgear/authgear-server/pkg/siteadmin"
	"github.com/authgear/authgear-server/pkg/util/pprofutil"
	"github.com/authgear/authgear-server/pkg/util/server"
	"github.com/authgear/authgear-server/pkg/util/signalutil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
	"github.com/authgear/authgear-server/pkg/version"
)

var logger = slogutil.NewLogger("authgear-portal")

type Controller struct {
	ServePortal    bool
	ServeSiteadmin bool
}

func (c *Controller) Start(ctx context.Context) {
	logger := logger.GetLogger(ctx)

	cfg, err := LoadConfigFromEnv(LoadConfigOptions{
		ServePortal:    c.ServePortal,
		ServeSiteadmin: c.ServeSiteadmin,
	})
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

	if c.ServePortal {
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
	}

	if c.ServeSiteadmin {
		// Shallow-copy the RootProvider so that the siteadmin server can use a
		// different AuthgearConfig (different AppID, Endpoint, etc.) without
		// affecting the portal server.
		//
		// A shallow copy is sufficient because:
		//   - Only AuthgearConfig needs to differ between portal and siteadmin.
		//   - All other fields (Database, RedisPool, ConfigSourceController, …)
		//     are pointers to shared infrastructure that both servers should reuse.
		//   - Dereferencing `p` copies the struct value, so overriding
		//     AuthgearConfig on the copy does not touch p.AuthgearConfig.
		siteadminProvider := *p
		siteadminProvider.AuthgearConfig = &cfg.SiteadminAuthgear
		specs = append(specs, server.NewSpec(ctx, &server.Spec{
			Name:          "authgear-portal-siteadmin",
			ListenAddress: cfg.SiteadminListenAddr,
			Handler:       siteadmin.NewRouter(&siteadminProvider),
		}))
		specs = append(specs, server.NewSpec(ctx, &server.Spec{
			Name:          "authgear-portal-siteadmin-internal",
			ListenAddress: cfg.SiteadminInternalListenAddr,
			Handler:       pprofutil.NewServeMux(),
		}))
	}

	signalutil.Start(ctx, specs...)
}
