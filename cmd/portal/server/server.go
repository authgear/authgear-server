package server

import (
	golog "log"

	"github.com/authgear/authgear-server/pkg/portal"
	"github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/server"
	"github.com/authgear/authgear-server/pkg/version"
)

type Controller struct {
	logger *log.Logger
}

func (c *Controller) Start() {
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		golog.Fatalf("failed to load server config: %s", err)
	}

	p, err := deps.NewRootProvider(cfg.EnvironmentConfig, cfg.ConfigSource, &cfg.Authgear, &cfg.AdminAPI, &cfg.App)
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
		Handler: portal.NewRouter(p, portal.StaticAssetConfig{
			ServingEnabled: cfg.StaticAsset.ServingEnabled,
			Directory:      cfg.StaticAsset.Dir,
		}),
	})
	server.Start(c.logger, specs)
}
