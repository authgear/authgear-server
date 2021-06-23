package configsource

import (
	"errors"
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

func LoadConfig(res *resource.Manager) (*config.Config, error) {
	result, err := res.Read(AppConfig, resource.EffectiveResource{})
	if errors.Is(err, resource.ErrResourceNotFound) {
		return nil, fmt.Errorf("missing '%s': %w", AuthgearYAML, err)
	} else if err != nil {
		return nil, err
	}
	appConfig := result.(*config.AppConfig)

	result, err = res.Read(SecretConfig, resource.EffectiveResource{})
	if errors.Is(err, resource.ErrResourceNotFound) {
		return nil, fmt.Errorf("missing '%s': %w", AuthgearSecretYAML, err)
	} else if err != nil {
		return nil, err
	}
	secretConfig := result.(*config.SecretConfig)

	var featureConfig *config.FeatureConfig
	result, err = res.Read(FeatureConfig, resource.EffectiveResource{})
	if errors.Is(err, resource.ErrResourceNotFound) {
		featureConfig = config.NewEffectiveDefaultFeatureConfig()
	} else if err != nil {
		return nil, err
	} else {
		featureConfig = result.(*config.FeatureConfig)
	}

	cfg := &config.Config{
		AppConfig:     appConfig,
		SecretConfig:  secretConfig,
		FeatureConfig: featureConfig,
	}
	if err = cfg.SecretConfig.Validate(cfg.AppConfig); err != nil {
		return nil, fmt.Errorf("invalid secret config: %w", err)
	}

	return cfg, nil
}
