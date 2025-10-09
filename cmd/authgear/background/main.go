package background

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/util/backgroundjob"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

type Controller struct{}

func (c *Controller) Start(ctx context.Context) {

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		err = fmt.Errorf("failed to load config: %w", err)
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
		err = fmt.Errorf("failed to setup server: %w", err)
		panic(err)
	}

	configSrcController := newConfigSourceController(p)
	err = configSrcController.Open(ctx)
	if err != nil {
		err = fmt.Errorf("cannot open configuration: %w", err)
		panic(err)
	}
	defer configSrcController.Close()

	runners := []*backgroundjob.Runner{
		newAccountDeletionRunner(ctx, p, configSrcController),
		newAccountAnonymizationRunner(ctx, p, configSrcController),
		newAccountStatusRunner(ctx, p, configSrcController),
	}
	backgroundjob.Main(ctx, runners)
}
