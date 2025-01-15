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

func (c *Controller) Start(ctx context.Context) {
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		golog.Fatalf("failed to load config: %v", err)
	}

	p, err := deps.NewBackgroundProvider(
		ctx,
		cfg.EnvironmentConfig,
		cfg.ConfigSource,
		cfg.CustomResourceDirectory,
	)
	if err != nil {
		golog.Fatalf("failed to setup server: %v", err)
	}

	// From now, we should use c.logger to log.
	c.logger = p.LoggerFactory.New("background")

	configSrcController := newConfigSourceController(p)
	err = configSrcController.Open(ctx)
	if err != nil {
		c.logger.WithError(err).Fatal("cannot open configuration")
	}
	defer configSrcController.Close()

	runners := []*backgroundjob.Runner{
		newAccountDeletionRunner(ctx, p, configSrcController),
		newAccountAnonymizationRunner(ctx, p, configSrcController),
	}
	backgroundjob.Main(ctx, c.logger, runners)
}
