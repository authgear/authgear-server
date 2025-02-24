package configsource

import (
	"context"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sync/atomic"

	"github.com/spf13/afero"
	"gopkg.in/fsnotify.v1"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

const LocalFSPlanName = "local-fs"

type LocalFSLogger struct{ *log.Logger }

func NewLocalFSLogger(lf *log.Factory) LocalFSLogger {
	return LocalFSLogger{lf.New("configsource-local-fs")}
}

type LocalFS struct {
	Logger        LocalFSLogger
	BaseResources *resource.Manager
	Config        *Config

	Fs      afero.Fs          `wire:"-"`
	config  atomic.Value      `wire:"-"`
	watcher *fsnotify.Watcher `wire:"-"`
	done    chan<- struct{}   `wire:"-"`
}

var _ ContextResolver = &LocalFS{}

func (s *LocalFS) Open(ctx context.Context) error {
	dir, err := filepath.Abs(s.Config.Directory)
	if err != nil {
		return err
	}

	s.Fs = afero.NewBasePathFs(afero.NewOsFs(), dir)
	appFs := &resource.LeveledAferoFs{Fs: s.Fs, FsLevel: resource.FsLevelApp}

	resources := s.BaseResources.Overlay(appFs)
	cfg, err := LoadConfig(ctx, resources)
	if err != nil {
		return err
	}

	s.config.Store(&config.AppContext{
		AppFs:     appFs,
		Resources: resources,
		Config:    cfg,
		PlanName:  LocalFSPlanName,
		Domains:   config.AppDomains{},
	})

	if s.Config.Watch {
		appConfigPath := path.Join(dir, AuthgearYAML)
		secretConfigPath := path.Join(dir, AuthgearSecretYAML)
		featureConfigPath := path.Join(dir, AuthgearFeatureYAML)

		s.watcher, err = fsnotify.NewWatcher()
		if err != nil {
			return err
		}

		done := make(chan struct{})
		s.done = done
		go s.watch(ctx, done)

		if err = s.watcher.Add(appConfigPath); err != nil {
			return err
		}
		if err = s.watcher.Add(secretConfigPath); err != nil {
			return err
		}
		if err = s.watcher.Add(featureConfigPath); err != nil {
			// watching feature config only works
			// when the authgear.features.yaml exists before the server starts
			if !os.IsNotExist(err) {
				return err
			}
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

func (s *LocalFS) watch(ctx context.Context, done <-chan struct{}) {
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
			if err := s.reload(ctx); err != nil {
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

func (s *LocalFS) reload(ctx context.Context) error {
	appCtx := s.config.Load().(*config.AppContext)

	newConfig, err := LoadConfig(ctx, appCtx.Resources)
	if err != nil {
		return err
	}

	appCtx = &config.AppContext{
		AppFs:     appCtx.AppFs,
		Resources: appCtx.Resources,
		Config:    newConfig,
		PlanName:  LocalFSPlanName,
		Domains:   config.AppDomains{},
	}
	s.config.Store(appCtx)
	return nil
}

func (s *LocalFS) AllAppIDs() ([]string, error) {
	ctx := s.config.Load().(*config.AppContext)
	appID := string(ctx.Config.AppConfig.ID)
	return []string{appID}, nil
}

func (s *LocalFS) ResolveAppID(ctx context.Context, r *http.Request) (appID string, err error) {
	// In single mode, appID is ignored.
	return
}

func (s *LocalFS) ResolveContext(ctx context.Context, _appID string, fn func(context.Context, *config.AppContext) error) error {
	// In single mode, appID is ignored.
	appCtx := s.config.Load().(*config.AppContext)
	ctx = config.WithAppContext(ctx, appCtx)
	return fn(ctx, appCtx)
}

func (s *LocalFS) ReloadApp(ctx context.Context, appID string) {
	// In single mode, appID is ignored.
	err := s.reload(ctx)
	if err != nil {
		s.Logger.
			WithError(err).
			Error("reload failed")
	}
}
