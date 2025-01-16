package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/userimport"
)

type End2End struct{}

func (c *End2End) ImportUsers(ctx context.Context, appID string, jsonPath string) error {
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		return err
	}
	cfg.ConfigSource = &configsource.Config{
		Type:  configsource.TypeDatabase,
		Watch: false,
	}

	p, err := deps.NewRootProvider(
		ctx,
		cfg.EnvironmentConfig,
		cfg.ConfigSource,
		cfg.CustomResourceDirectory,
	)
	if err != nil {
		return err
	}

	configSrcController := newConfigSourceController(p)
	err = configSrcController.Open(ctx)
	if err != nil {
		return err
	}
	defer configSrcController.Close()

	appCtx, err := configSrcController.ResolveContext(ctx, appID)
	if err != nil {
		return err
	}

	appProvider := p.NewAppProvider(ctx, appCtx)

	userImport := newUserImport(appProvider)

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

	res := userImport.ImportRecords(ctx, &request)
	if res.Summary.Failed > 0 {
		return fmt.Errorf("failed to import %d records due to %v", res.Summary.Failed, res.Details)
	}

	return nil
}
