package server

import (
	"context"
	golog "log"

	"github.com/authgear/authgear-server/pkg/images"
	"github.com/authgear/authgear-server/pkg/images/deps"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/server"
	"github.com/authgear/authgear-server/pkg/util/vipsutil"
	"github.com/authgear/authgear-server/pkg/version"
)

type Controller struct {
	logger *log.Logger
}

func (c *Controller) Start() {
	vipsutil.LibvipsInit()

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		golog.Fatalf("failed to load server config: %s", err)
	}

	p, err := deps.NewRootProvider(*cfg.EnvironmentConfig, cfg.ObjectStore)
	if err != nil {
		golog.Fatalf("failed to setup server: %s", err)
	}

	configSrcController := newConfigSourceController(p, context.Background())
	err = configSrcController.Open()
	if err != nil {
		c.logger.WithError(err).Fatal("cannot open configuration")
	}
	defer configSrcController.Close()

	// From now, we should use c.logger to log.
	c.logger = p.LoggerFactory.New("authgear-images")
	c.logger.Infof("authgear (version %s)", version.Version)

	var specs []server.Spec
	specs = append(specs, server.Spec{
		Name:          "images server",
		ListenAddress: cfg.ListenAddr,
		Handler:       images.NewRouter(p, configSrcController.GetConfigSource()),
	})
	server.Start(c.logger, specs)
}
