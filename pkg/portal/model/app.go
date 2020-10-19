package model

import (
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/portal/resources"
)

type App struct {
	ID      string
	Context *config.AppContext
}

func (a *App) LoadRawAppConfig() (*config.AppConfig, error) {
	files, err := configsource.AppConfig.ReadResource(a.Context.AppFs)
	if err != nil {
		return nil, err
	}

	var cfg *config.AppConfig
	if err := yaml.Unmarshal(files[0].Data, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (a *App) LoadRawSecretConfig() (*config.SecretConfig, error) {
	files, err := configsource.SecretConfig.ReadResource(a.Context.AppFs)
	if err != nil {
		return nil, err
	}

	var cfg *config.SecretConfig
	if err := yaml.Unmarshal(files[0].Data, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

type AppResource struct {
	resources.Resource
	Context *config.AppContext
}
