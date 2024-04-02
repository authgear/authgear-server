package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"dario.cat/mergo"
	cp "github.com/otiai10/copy"
	"gopkg.in/yaml.v2"

	"github.com/authgear/authgear-server/cmd/portal/internal"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
)

type End2End struct {
	Context context.Context
}

type NoopTaskQueue struct{}

func (q NoopTaskQueue) Enqueue(param task.Param) {
}

var BuiltInConfigSourceDir = "./var"

func (c *End2End) CreateApp(appID string, baseConfigSourceDir string, override string) error {
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		return err
	}

	configSourceDir, err := c.createTempConfigSource(appID, baseConfigSourceDir, override)
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

func (c *End2End) createTempConfigSource(appID string, baseConfigSource string, overrideYAML string) (string, error) {
	tempAppDir, err := os.MkdirTemp("", "e2e-")
	if err != nil {
		return "", err
	}

	err = cp.Copy(BuiltInConfigSourceDir, tempAppDir)
	if err != nil {
		return "", err
	}

	baseConfigSourceInfo, err := os.Stat(baseConfigSource)
	if err != nil {
		return "", err
	}

	dest := tempAppDir
	if !baseConfigSourceInfo.IsDir() {
		dest = filepath.Join(tempAppDir, baseConfigSourceInfo.Name())
	}
	err = cp.Copy(baseConfigSource, dest)
	if err != nil {
		return "", err
	}

	authgearYAMLPath := filepath.Join(tempAppDir, configsource.AuthgearYAML)
	authgearYAML, err := os.ReadFile(authgearYAMLPath)
	if err != nil {
		return "", err
	}

	cfg, err := config.Parse([]byte(authgearYAML))
	if err != nil {
		return "", err
	}

	var overrideCfg config.AppConfig
	jsonData, err := yaml.Marshal([]byte(overrideYAML))
	if err != nil {
		return "", err
	}

	decoder := json.NewDecoder(bytes.NewReader(jsonData))
	err = decoder.Decode(&overrideCfg)
	if err != nil {
		return "", err
	}

	err = mergo.Merge(cfg, &overrideCfg, mergo.WithOverride)
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
