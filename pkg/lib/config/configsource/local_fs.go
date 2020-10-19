package configsource

import (
	"net/http"
	"path/filepath"
	"sync/atomic"

	"github.com/spf13/afero"
	"gopkg.in/fsnotify.v1"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

type LocalFSLogger struct{ *log.Logger }

func NewLocalFSLogger(lf *log.Factory) LocalFSLogger {
	return LocalFSLogger{lf.New("configsource-local-fs")}
}

type LocalFS struct {
	Logger        LocalFSLogger
	BaseResources *resource.Manager
	Config        *Config

	Fs               afero.Fs          `wire:"-"`
	appConfigPath    string            `wire:"-"`
	secretConfigPath string            `wire:"-"`
	config           atomic.Value      `wire:"-"`
	watcher          *fsnotify.Watcher `wire:"-"`
	done             chan<- struct{}   `wire:"-"`
}

func (s *LocalFS) Open() error {
	dir, err := filepath.Abs(s.Config.Directory)
	if err != nil {
		return err
	}

	s.Fs = afero.NewBasePathFs(afero.NewOsFs(), dir)
	appFs := &resource.AferoFs{Fs: s.Fs}

	resources := s.BaseResources.Overlay(appFs)
	cfg, err := LoadConfig(resources)
	if err != nil {
		return err
	}

	s.config.Store(&config.AppContext{
		AppFs:     appFs,
		Resources: resources,
		Config:    cfg,
	})

	if s.Config.Watch {
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

func (s *LocalFS) Close() error {
	if s.watcher != nil {
		close(s.done)
		return s.watcher.Close()
	}
	return nil
}

func (s *LocalFS) watch(done <-chan struct{}) {
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

			if err := s.reload(); err != nil {
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

func (s *LocalFS) reload() error {
	appCtx := s.config.Load().(*config.AppContext)

	newConfig, err := LoadConfig(appCtx.Resources)
	if err != nil {
		return err
	}

	appCtx = &config.AppContext{
		AppFs:     appCtx.AppFs,
		Resources: appCtx.Resources,
		Config:    newConfig,
	}
	s.config.Store(appCtx)
	return nil
}

func (s *LocalFS) AllAppIDs() ([]string, error) {
	ctx := s.config.Load().(*config.AppContext)
	appID := string(ctx.Config.AppConfig.ID)
	return []string{appID}, nil
}

func (s *LocalFS) ResolveAppID(r *http.Request) (appID string, err error) {
	// In single mode, appID is ignored.
	return
}

func (s *LocalFS) ResolveContext(_appID string) (*config.AppContext, error) {
	// In single mode, appID is ignored.
	ctx := s.config.Load().(*config.AppContext)
	return ctx, nil
}

func (s *LocalFS) ReloadApp(appID string) {
	// In single mode, appID is ignored.
	err := s.reload()
	if err != nil {
		s.Logger.
			WithError(err).
			Error("reload failed")
	}
}
