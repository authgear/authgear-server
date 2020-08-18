package server

import (
	golog "log"

	"github.com/authgear/authgear-server/pkg/portal"
	"github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/server"
	"github.com/authgear/authgear-server/pkg/version"
)

type Controller struct {
	logger *log.Logger
}

func (c *Controller) Start() {
	cfg, err := config.LoadServerConfigFromEnv()
	if err != nil {
		golog.Fatalf("failed to load server config: %s", err)
	}

	p, err := deps.NewRootProvider(cfg)
	if err != nil {
		golog.Fatalf("failed to setup server: %s", err)
	}

	// From now, we should use c.logger to log.
	c.logger = p.LoggerFactory.New("authgear-portal")

	c.logger.Infof("authgear-portal (version %s)", version.Version)
	if cfg.DevMode {
		c.logger.Warn("development mode is ON - do not use in production")
	}

	var specs []server.Spec

	if cfg.DevMode {
		specs = append(specs, server.Spec{
			Name:          "public server",
			ListenAddress: cfg.ListenAddr,
			HTTPS:         true,
			CertFilePath:  cfg.TLSCertFilePath,
			KeyFilePath:   cfg.TLSKeyFilePath,
			Handler:       portal.NewRouter(p),
		})
	} else {
		specs = append(specs, server.Spec{
			Name:          "public server",
			ListenAddress: cfg.ListenAddr,
			Handler:       portal.NewRouter(p),
		})
	}

	server.Start(c.logger, specs)
}
