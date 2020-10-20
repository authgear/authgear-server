package configsource

import (
	"errors"
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

func LoadConfig(res *resource.Manager) (*config.Config, error) {
	appConfigFile, err := res.Read(AppConfig, nil)
	if errors.Is(err, resource.ErrResourceNotFound) {
		return nil, fmt.Errorf("missing '%s': %w", AuthgearYAML, err)
	} else if err != nil {
		return nil, err
	}
	appConfig, err := AppConfig.Parse(appConfigFile)
	if err != nil {
		return nil, err
	}

	secretConfigFile, err := res.Read(SecretConfig, nil)
	if errors.Is(err, resource.ErrResourceNotFound) {
		return nil, fmt.Errorf("missing '%s': %w", AuthgearSecretYAML, err)
	} else if err != nil {
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
