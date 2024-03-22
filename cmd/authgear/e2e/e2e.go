package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
	"github.com/authgear/authgear-server/pkg/lib/userimport"
)

type End2End struct {
	Context context.Context

	ConfigSource configsource.Config
}

type NoopTaskQueue struct{}

func (q NoopTaskQueue) Enqueue(param task.Param) {
}

func (c *End2End) CreateUserFixtures(jsonPath string) error {
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		return err
	}
	cfg.ConfigSource = &c.ConfigSource

	taskQueueFactory := deps.TaskQueueFactory(func(provider *deps.AppProvider) task.Queue {
		return NoopTaskQueue{}
	})

	p, err := deps.NewRootProvider(
		config.MainListenAddr(cfg.MainListenAddr),
		cfg.EnvironmentConfig,
		cfg.ConfigSource,
		cfg.BuiltinResourceDirectory,
		cfg.CustomResourceDirectory,
		taskQueueFactory,
	)
	if err != nil {
		return err
	}

	configSrcController := newConfigSourceController(p, context.Background())
	err = configSrcController.Open()
	if err != nil {
		return err
	}
	defer configSrcController.Close()

	appCtx, err := configSrcController.ResolveContext("")
	if err != nil {
		return err
	}

	appProvider := p.NewAppProvider(c.Context, appCtx)

	userImport := newUserImport(appProvider, c.Context)

	jsoFile, err := os.Open(jsonPath)
	if err != nil {
		return err
	}
	defer jsoFile.Close()

	var request userimport.Request
	err = json.NewDecoder(jsoFile).Decode(&request)
	if err != nil {
		return err
	}

	res := userImport.ImportRecords(c.Context, &request)
	if res.Summary.Failed > 0 {
		return fmt.Errorf("failed to import %d records due to %v", res.Summary.Failed, res.Details)
	}

	return nil
}
