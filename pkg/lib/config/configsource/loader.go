package configsource

import (
	"context"
	"errors"
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

func LoadConfig(ctx context.Context, res *resource.Manager) (*config.Config, error) {

	result, err := res.Read(ctx, AppConfig, resource.EffectiveResource{})
	if errors.Is(err, resource.ErrResourceNotFound) {
		return nil, fmt.Errorf("missing '%s': %w", AuthgearYAML, err)
	} else if err != nil {
		return nil, errors.Join(errors.New("failed to read app config"), err)
	}
	appConfig := result.(*config.AppConfig)

	result, err = res.Read(ctx, SecretConfig, resource.EffectiveResource{})
	if errors.Is(err, resource.ErrResourceNotFound) {
		return nil, fmt.Errorf("missing '%s': %w", AuthgearSecretYAML, err)
	} else if err != nil {
		return nil, errors.Join(errors.New("failed to read secret config"), err)
	}
	secretConfig := result.(*config.SecretConfig)

	var featureConfig *config.FeatureConfig
	result, err = res.Read(ctx, FeatureConfig, resource.EffectiveResource{})
	if errors.Is(err, resource.ErrResourceNotFound) {
		featureConfig = config.NewEffectiveDefaultFeatureConfig()
	} else if err != nil {
		return nil, errors.Join(errors.New("failed to read feature config"), err)
	} else {
		featureConfig = result.(*config.FeatureConfig)
	}

	cfg := &config.Config{
		AppConfig:     appConfig,
		SecretConfig:  secretConfig,
		FeatureConfig: featureConfig,
	}
	if err = cfg.SecretConfig.Validate(ctx, cfg.AppConfig); err != nil {
		return nil, fmt.Errorf("invalid secret config: %w", err)
	}

	return cfg, nil
}
