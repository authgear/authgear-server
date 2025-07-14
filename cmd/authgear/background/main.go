package background

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/util/backgroundjob"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var logger = slogutil.NewLogger("background")

type Controller struct{}

func (c *Controller) Start(ctx context.Context) {
	logger := logger.GetLogger(ctx)

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		logger.WithError(err).Error(ctx, "failed to load config")
		panic(err)
	}

	ctx = slogutil.Setup(ctx)

	ctx, p, err := deps.NewBackgroundProvider(
		ctx,
		cfg.EnvironmentConfig,
		cfg.ConfigSource,
		cfg.CustomResourceDirectory,
	)
	if err != nil {
		logger.WithError(err).Error(ctx, "failed to setup server")
		panic(err)
	}

	configSrcController := newConfigSourceController(p)
	err = configSrcController.Open(ctx)
	if err != nil {
		logger.WithError(err).Error(ctx, "cannot open configuration")
		panic(err)
	}
	defer configSrcController.Close()

	runners := []*backgroundjob.Runner{
		newAccountDeletionRunner(ctx, p, configSrcController),
		newAccountAnonymizationRunner(ctx, p, configSrcController),
	}
	backgroundjob.Main(ctx, runners)
}
