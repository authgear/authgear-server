package background

import (
	"context"
	golog "log"

	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/util/backgroundjob"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type Controller struct {
	logger *log.Logger
}

func (c *Controller) Start() {
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		golog.Fatalf("failed to load config: %v", err)
	}

	p, err := deps.NewBackgroundProvider(
		cfg.EnvironmentConfig,
		cfg.ConfigSource,
		cfg.BuiltinResourceDirectory,
		cfg.CustomResourceDirectory,
	)
	if err != nil {
		golog.Fatalf("failed to setup server: %v", err)
	}

	// From now, we should use c.logger to log.
	c.logger = p.LoggerFactory.New("background")

	configSrcController := newConfigSourceController(p, context.Background())
	err = configSrcController.Open()
	if err != nil {
		c.logger.WithError(err).Fatal("cannot open configuration")
	}
	defer configSrcController.Close()

	// FIXME: configure runners
	var runners []*backgroundjob.Runner
	backgroundjob.Main(c.logger, runners)
}
