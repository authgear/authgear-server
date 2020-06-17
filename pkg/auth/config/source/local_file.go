package source

import (
	"context"
	"fmt"
	"github.com/skygeario/skygear-server/pkg/auth/config"
	"io/ioutil"
	"net/http"
)

type LocalFile struct {
	serverConfig *config.ServerConfig
	config       *config.Config
}

func NewLocalFile(cfg *config.ServerConfig) *LocalFile {
	return &LocalFile{
		serverConfig: cfg,
	}
}

func (s *LocalFile) Start() error {
	appConfigYAML, err := ioutil.ReadFile(s.serverConfig.ConfigSource.AppConfigPath)
	if err != nil {
		return fmt.Errorf("cannot read app config file: %w", err)
	}
	appConfig, err := config.Parse(appConfigYAML)
	if err != nil {
		return fmt.Errorf("cannot parse app config: %w", err)
	}

	secretConfigYAML, err := ioutil.ReadFile(s.serverConfig.ConfigSource.SecretConfigPath)
	if err != nil {
		return fmt.Errorf("cannot read secret config file: %w", err)
	}
	secretConfig, err := config.ParseSecret(secretConfigYAML)
	if err != nil {
		return fmt.Errorf("cannot parse secret config: %w", err)
	}

	if err = secretConfig.Validate(appConfig); err != nil {
		return fmt.Errorf("invalid secret config: %w", err)
	}

	s.config = &config.Config{
		AppConfig:    appConfig,
		SecretConfig: secretConfig,
	}
	return nil
}

func (s *LocalFile) Shutdown() error {
	return nil
}

func (s *LocalFile) ProvideConfig(ctx context.Context, r *http.Request) (*config.Config, error) {
	return s.config, nil
}
