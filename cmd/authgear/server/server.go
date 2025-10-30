package server

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/authgear/authgear-server/pkg/admin"
	"github.com/authgear/authgear-server/pkg/auth"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	infraredisqueue "github.com/authgear/authgear-server/pkg/lib/infra/redisqueue"
	"github.com/authgear/authgear-server/pkg/redisqueue"
	"github.com/authgear/authgear-server/pkg/resolver"
	"github.com/authgear/authgear-server/pkg/util/pprofutil"
	"github.com/authgear/authgear-server/pkg/util/server"
	"github.com/authgear/authgear-server/pkg/util/signalutil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
	"github.com/authgear/authgear-server/pkg/version"
)

var logger = slogutil.NewLogger("server")

type Controller struct {
	ServeMain     bool
	ServeResolver bool
	ServeAdmin    bool
}

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
		cfg.ConfigSource,
		cfg.CustomResourceDirectory,
	)
	if err != nil {
		err = fmt.Errorf("failed to setup server: %w", err)
		panic(err)
	}

	logger.Info(ctx, "authgear version", slog.String("version", version.Version))
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

	var specs []signalutil.Daemon

	if c.ServeMain {
		u, err := server.ParseListenAddress(cfg.MainListenAddr)
		if err != nil {
			err = fmt.Errorf("failed to parse main server listen address: %w", err)
			panic(err)
		}

		spec := &server.Spec{
			Name:          "authgear-main",
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

		specs = append(specs, server.NewSpec(ctx, spec))

		// Set up internal server.
		u, err = server.ParseListenAddress(cfg.MainInteralListenAddr)
		if err != nil {
			err = fmt.Errorf("failed to parse main server internal listen address: %w", err)
			panic(err)
		}
		specs = append(specs, server.NewSpec(ctx, &server.Spec{
			Name:          "authgear-main-internal",
			ListenAddress: u.Host,
			Handler:       pprofutil.NewServeMux(),
		}))

		specs = append(specs, redisqueue.NewConsumer(
			ctx,
			infraredisqueue.QueueUserReindex,
			cfg.RateLimits.TaskUserReindex,
			p,
			configSrcController,
			redisqueue.UserReindex,
		))
	}

	if c.ServeResolver {
		u, err := server.ParseListenAddress(cfg.ResolverListenAddr)
		if err != nil {
			err = fmt.Errorf("failed to parse resolver server listen address: %w", err)
			panic(err)
		}

		specs = append(specs, server.NewSpec(ctx, &server.Spec{
			Name:          "authgear-resolver",
			ListenAddress: u.Host,
			Handler: resolver.NewRouter(
				p,
				configSrcController.GetConfigSource(),
			),
		}))

		// Set up internal server.
		u, err = server.ParseListenAddress(cfg.ResolverInternalListenAddr)
		if err != nil {
			err = fmt.Errorf("failed to parse resolver internal server listen address: %w", err)
			panic(err)
		}

		specs = append(specs, server.NewSpec(ctx, &server.Spec{
			Name:          "authgear-resolver-internal",
			ListenAddress: u.Host,
			Handler:       pprofutil.NewServeMux(),
		}))
	}

	if c.ServeAdmin {
		u, err := server.ParseListenAddress(cfg.AdminListenAddr)
		if err != nil {
			err = fmt.Errorf("failed to parse admin API server listen address: %w", err)
			panic(err)
		}

		specs = append(specs, server.NewSpec(ctx, &server.Spec{
			Name:          "authgear-admin",
			ListenAddress: u.Host,
			Handler: admin.NewRouter(
				p,
				configSrcController.GetConfigSource(),
				cfg.AdminAPIAuth,
			),
		}))

		u, err = server.ParseListenAddress(cfg.AdminInternalListenAddr)
		if err != nil {
			err = fmt.Errorf("failed to parse admin API internal server listen address: %w", err)
			panic(err)
		}

		specs = append(specs, server.NewSpec(ctx, &server.Spec{
			Name:          "authgear-admin-internal",
			ListenAddress: u.Host,
			Handler:       pprofutil.NewServeMux(),
		}))

		specs = append(specs, redisqueue.NewConsumer(
			ctx,
			infraredisqueue.QueueUserImport,
			cfg.RateLimits.TaskUserImport,
			p,
			configSrcController,
			redisqueue.UserImport,
		))

		specs = append(specs, redisqueue.NewConsumer(
			ctx,
			infraredisqueue.QueueUserExport,
			cfg.RateLimits.TaskUserExport,
			p,
			configSrcController,
			redisqueue.UserExport,
		))
	}

	signalutil.Start(ctx, specs...)
}
