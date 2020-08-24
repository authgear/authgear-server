package configsource

import (
	"fmt"
	"io/ioutil"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/fs"
)

const (
	AuthgearYAML       = "authgear.yaml"
	AuthgearSecretYAML = "authgear.secrets.yaml"
)

func loadConfig(fs fs.Fs) (*config.Config, error) {
	appConfigFile, err := fs.Open(AuthgearYAML)
	if err != nil {
		return nil, fmt.Errorf("cannot open app config file: %w", err)
	}
	appConfigYAML, err := ioutil.ReadAll(appConfigFile)
	if err != nil {
		return nil, fmt.Errorf("cannot read app config file: %w", err)
	}
	appConfig, err := config.Parse(appConfigYAML)
	if err != nil {
		return nil, fmt.Errorf("cannot parse app config: %w", err)
	}

	secretConfigFile, err := fs.Open(AuthgearSecretYAML)
	if err != nil {
		return nil, fmt.Errorf("cannot open secret config file: %w", err)
	}
	secretConfigYAML, err := ioutil.ReadAll(secretConfigFile)
	if err != nil {
		return nil, fmt.Errorf("cannot read secret config file: %w", err)
	}
	secretConfig, err := config.ParseSecret(secretConfigYAML)
	if err != nil {
		return nil, fmt.Errorf("cannot parse secret config: %w", err)
	}

	if err = secretConfig.Validate(appConfig); err != nil {
		return nil, fmt.Errorf("invalid secret config: %w", err)
	}

	return &config.Config{
		AppConfig:    appConfig,
		SecretConfig: secretConfig,
	}, nil
}
