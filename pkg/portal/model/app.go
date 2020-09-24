package model

import (
	"io/ioutil"

	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
)

type App struct {
	ID      string
	Context *config.AppContext
}

func (a *App) LoadFile(path string) ([]byte, error) {
	file, err := a.Context.Fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (a *App) LoadAppConfigFile() (*config.AppConfig, error) {
	data, err := a.LoadFile(configsource.AuthgearYAML)
	if err != nil {
		return nil, err
	}
	var cfg *config.AppConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (a *App) LoadSecretConfigFile() (*config.SecretConfig, error) {
	data, err := a.LoadFile(configsource.AuthgearSecretYAML)
	if err != nil {
		return nil, err
	}
	var cfg *config.SecretConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

type AppConfigFile struct {
	Path    string
	Content string
}
