package configsource

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

func loadConfig(res *resource.Manager) (*config.Config, error) {
	appConfigFile, err := res.Read(AppConfig, nil)
	if err != nil {
		return nil, err
	}
	appConfig, err := AppConfig.Parse(appConfigFile)
	if err != nil {
		return nil, err
	}

	secretConfigFile, err := res.Read(SecretConfig, nil)
	if err != nil {
		return nil, err
	}
	secretConfig, err := SecretConfig.Parse(secretConfigFile)
	if err != nil {
		return nil, err
	}

	cfg := &config.Config{
		AppConfig:    appConfig.(*config.AppConfig),
		SecretConfig: secretConfig.(*config.SecretConfig),
	}
	if err = cfg.SecretConfig.Validate(cfg.AppConfig); err != nil {
		return nil, fmt.Errorf("invalid secret config: %w", err)
	}

	return cfg, nil
}
