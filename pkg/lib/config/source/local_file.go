package source

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"sync/atomic"

	"gopkg.in/fsnotify.v1"

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

	appConfigPath    string            `wire:"-"`
	secretConfigPath string            `wire:"-"`
	config           atomic.Value      `wire:"-"`
	watcher          *fsnotify.Watcher `wire:"-"`
	done             chan<- struct{}   `wire:"-"`
}

func (s *LocalFile) Open() error {
	dir, err := filepath.Abs(s.ServerConfig.ConfigSource.ConfigDirectory)
	if err != nil {
		return err
	}

	s.appConfigPath = filepath.Join(dir, AuthgearYAML)
	appConfigYAML, err := ioutil.ReadFile(s.appConfigPath)
	if err != nil {
		return fmt.Errorf("cannot read app config file: %w", err)
	}
	appConfig, err := config.Parse(appConfigYAML)
	if err != nil {
		return fmt.Errorf("cannot parse app config: %w", err)
	}

	s.secretConfigPath = filepath.Join(dir, AuthgearSecretYAML)
	secretConfigYAML, err := ioutil.ReadFile(s.secretConfigPath)
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

	s.config.Store(&config.Config{
		BaseDirectory: dir,
		AppConfig:     appConfig,
		SecretConfig:  secretConfig,
	})

	if s.ServerConfig.ConfigSource.Watch {
		s.watcher, err = fsnotify.NewWatcher()
		if err != nil {
			return err
		}

		done := make(chan struct{})
		s.done = done
		go s.watch(done)

		if err = s.watcher.Add(s.appConfigPath); err != nil {
			return err
		}
		if err = s.watcher.Add(s.secretConfigPath); err != nil {
			return err
		}
	}
	return nil
}

func (s *LocalFile) Close() error {
	if s.watcher != nil {
		close(s.done)
		return s.watcher.Close()
	}
	return nil
}

func (s *LocalFile) watch(done <-chan struct{}) {
	for {
		select {
		case event, ok := <-s.watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write != fsnotify.Write {
				break
			}
			s.Logger.
				WithField("file", event.Name).
				Info("change detected, reloading...")

			if err := s.reload(event.Name); err != nil {
				s.Logger.
					WithError(err).
					WithField("file", event.Name).
					Error("reload failed")
			}

		case err, ok := <-s.watcher.Errors:
			if !ok {
				return
			}
			s.Logger.WithError(err).Fatal("Watcher failed")

		case <-done:
			return
		}
	}
}

func (s *LocalFile) reload(filename string) error {
	newConfig := *s.config.Load().(*config.Config)

	switch filename {
	case s.appConfigPath:
		appConfigYAML, err := ioutil.ReadFile(s.appConfigPath)
		if err != nil {
			return fmt.Errorf("cannot read app config file: %w", err)
		}
		newConfig.AppConfig, err = config.Parse(appConfigYAML)
		if err != nil {
			return fmt.Errorf("cannot parse app config: %w", err)
		}

	case s.secretConfigPath:
		secretConfigYAML, err := ioutil.ReadFile(s.secretConfigPath)
		if err != nil {
			return fmt.Errorf("cannot read secret config file: %w", err)
		}
		newConfig.SecretConfig, err = config.ParseSecret(secretConfigYAML)
		if err != nil {
			return fmt.Errorf("cannot parse secret config: %w", err)
		}
	}

	if err := newConfig.SecretConfig.Validate(newConfig.AppConfig); err != nil {
		return fmt.Errorf("invalid secret config: %w", err)
	}

	s.config.Store(&newConfig)
	return nil
}

func (s *LocalFile) ProvideConfig(ctx context.Context, r *http.Request, server ServerType) (*config.Config, error) {
	cfg := s.config.Load().(*config.Config)

	if s.ServerConfig.DevMode {
		// Accept all hosts under development mode
		return cfg, nil
	}

	var acceptHosts []string
	switch server {
	case ServerTypeMain, ServerTypeResolver:
		acceptHosts = cfg.AppConfig.HTTP.Hosts
	case ServerTypeAdminAPI:
		acceptHosts = cfg.AppConfig.HTTP.AdminHosts
	}

	host := httputil.GetHost(r, s.ServerConfig.TrustProxy)
	for _, h := range acceptHosts {
		if h == host {
			return cfg, nil
		}
	}
	s.Logger.Debugf("expected host %v, got %s", acceptHosts, host)
	return nil, fmt.Errorf("request host is not valid: %w", ErrAppNotFound)
}
