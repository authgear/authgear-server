package source

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type LocalFileLogger struct{ *log.Logger }

func NewLocalFileLogger(lf *log.Factory) LocalFileLogger {
	return LocalFileLogger{lf.New("local-file-config")}
}

type LocalFile struct {
	Logger       LocalFileLogger
	ServerConfig *config.ServerConfig

	config *config.Config `wire:"-"`
}

func (s *LocalFile) Open() error {
	dir, err := filepath.Abs(s.ServerConfig.ConfigSource.ConfigDirectory)
	if err != nil {
		return err
	}

	appConfigYAML, err := ioutil.ReadFile(filepath.Join(dir, AuthgearYAML))
	if err != nil {
		return fmt.Errorf("cannot read app config file: %w", err)
	}
	appConfig, err := config.Parse(appConfigYAML)
	if err != nil {
		return fmt.Errorf("cannot parse app config: %w", err)
	}

	secretConfigYAML, err := ioutil.ReadFile(filepath.Join(dir, AuthgearSecretYAML))
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
		BaseDirectory: dir,
		AppConfig:     appConfig,
		SecretConfig:  secretConfig,
	}
	return nil
}

func (s *LocalFile) Close() error {
	return nil
}

func (s *LocalFile) ProvideConfig(ctx context.Context, r *http.Request, server ServerType) (*config.Config, error) {
	if s.ServerConfig.DevMode {
		// Accept all hosts under development mode
		return s.config, nil
	}

	var acceptHosts []string
	switch server {
	case ServerTypeMain, ServerTypeResolver:
		acceptHosts = s.config.AppConfig.HTTP.Hosts
	case ServerTypeAdminAPI:
		acceptHosts = s.config.AppConfig.HTTP.AdminHosts
	}

	host := httputil.GetHost(r, s.ServerConfig.TrustProxy)
	for _, h := range acceptHosts {
		if h == host {
			return s.config, nil
		}
	}
	s.Logger.Debugf("expected host %v, got %s", acceptHosts, host)
	return nil, fmt.Errorf("request host is not valid: %w", ErrAppNotFound)
}
