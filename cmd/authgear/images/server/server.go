package server

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/authgear/authgear-server/pkg/images"
	"github.com/authgear/authgear-server/pkg/images/deps"
	"github.com/authgear/authgear-server/pkg/util/pprofutil"
	"github.com/authgear/authgear-server/pkg/util/server"
	"github.com/authgear/authgear-server/pkg/util/signalutil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
	"github.com/authgear/authgear-server/pkg/util/vipsutil"
	"github.com/authgear/authgear-server/pkg/version"
)

var logger = slogutil.NewLogger("authgear-images")

type Controller struct{}

func (c *Controller) Start(ctx context.Context) {
	logger := logger.GetLogger(ctx)

	vipsutil.LibvipsInit()

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		err = fmt.Errorf("failed to load server config: %w", err)
		panic(err)
	}

	ctx, p, err := deps.NewRootProvider(ctx, *cfg.EnvironmentConfig, cfg.ObjectStore)
	if err != nil {
		err = fmt.Errorf("failed to initialize dependencies: %w", err)
		panic(err)
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

	// From now, we should use c.logger to log.
	logger.Info(ctx, "authgear version", slog.String("version", version.Version))

	var specs []signalutil.Daemon
	specs = append(specs, server.NewSpec(ctx, &server.Spec{
		Name:          "authgear-images",
		ListenAddress: cfg.ListenAddr,
		Handler:       images.NewRouter(p, configSrcController.GetConfigSource()),
	}))
	specs = append(specs, server.NewSpec(ctx, &server.Spec{
		Name:          "authgear-images-internal",
		ListenAddress: cfg.InternalListenAddr,
		Handler:       pprofutil.NewServeMux(),
	}))
	signalutil.Start(ctx, specs...)
}
