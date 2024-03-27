package e2e

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"io/fs"

	"github.com/authgear/authgear-server/cmd/portal/internal"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
	cp "github.com/otiai10/copy"
)

type End2End struct {
	Context context.Context
}

type NoopTaskQueue struct{}

func (q NoopTaskQueue) Enqueue(param task.Param) {
}

func (c *End2End) CreateApp(appID string, baseConfigSourceDir string) error {
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		return err
	}

	configSourceDir, err := c.createTempConfigSource(appID, baseConfigSourceDir)
	if err != nil {
		return err
	}

	err = internal.Create(&internal.CreateOptions{
		DatabaseURL:    cfg.GlobalDatabase.DatabaseURL,
		DatabaseSchema: cfg.GlobalDatabase.DatabaseSchema,
		ResourceDir:    configSourceDir,
	})
	if err != nil {
		return err
	}

	err = internal.CreateDefaultDomain(internal.CreateDefaultDomainOptions{
		DatabaseURL:         cfg.GlobalDatabase.DatabaseURL,
		DatabaseSchema:      cfg.GlobalDatabase.DatabaseSchema,
		DefaultDomainSuffix: ".portal.localhost",
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *End2End) createTempConfigSource(appID string, baseConfigSourceDir string) (string, error) {
	tempAppDir, err := os.MkdirTemp("", "e2e-")
	if err != nil {
		return "", err
	}

	err = cp.Copy(baseConfigSourceDir, tempAppDir)
	if err != nil {
		return "", err
	}

	authgearYAMLPath := filepath.Join(tempAppDir, configsource.AuthgearYAML)
	authgearYAML, err := os.ReadFile(authgearYAMLPath)
	if err != nil {
		return "", err
	}

	cfg, err := config.Parse(authgearYAML)
	if err != nil {
		return "", err
	}

	cfg.ID = config.AppID(appID)
	cfg.HTTP.PublicOrigin = fmt.Sprintf("http://%s.portal.localhost:4000", appID)

	newAuthgearYAML, err := config.Export(cfg)
	if err != nil {
		return "", err
	}

	err = os.WriteFile(authgearYAMLPath, newAuthgearYAML, fs.FileMode(0644))
	if err != nil {
		return "", err
	}

	return tempAppDir, nil
}
