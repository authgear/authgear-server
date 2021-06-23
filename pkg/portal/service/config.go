package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/afero"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

var ErrDuplicatedAppID = apierrors.AlreadyExists.WithReason("DuplicatedAppID").
	New("duplicated app ID")

var ErrGetStaticAppIDsNotSupported = errors.New("only local FS config source can get static app ID")

type IngressTemplateData struct {
	AppID         string
	DomainID      string
	IsCustom      bool
	Host          string
	TLSSecretName string
}

type ConfigServiceLogger struct{ *log.Logger }

func NewConfigServiceLogger(lf *log.Factory) ConfigServiceLogger {
	return ConfigServiceLogger{lf.New("config-service")}
}

type CreateAppOptions struct {
	AppID     string
	Resources map[string][]byte
	PlanName  string
}

type ConfigService struct {
	Context              context.Context
	Logger               ConfigServiceLogger
	AppConfig            *portalconfig.AppConfig
	Controller           *configsource.Controller
	ConfigSource         *configsource.ConfigSource
	DomainImplementation portalconfig.DomainImplementationType
	Kubernetes           *Kubernetes
}

func (s *ConfigService) ResolveContext(appID string) (*config.AppContext, error) {
	return s.ConfigSource.ContextResolver.ResolveContext(appID)
}

func (s *ConfigService) GetStaticAppIDs() ([]string, error) {
	switch src := s.Controller.Handle.(type) {
	case *configsource.Database:
		return nil, ErrGetStaticAppIDsNotSupported
	case *configsource.LocalFS:
		return src.AllAppIDs()
	default:
		return nil, errors.New("unsupported configuration source")
	}
}

func (s *ConfigService) Create(opts *CreateAppOptions) error {
	switch src := s.Controller.Handle.(type) {
	case *configsource.Database:
		err := s.createDatabase(src, opts)
		if err != nil {
			return err
		}
	case *configsource.LocalFS:
		return apierrors.NewForbidden("cannot create app for local FS")

	default:
		return errors.New("unsupported configuration source")
	}
	return nil
}

func (s *ConfigService) UpdateResources(appID string, files []*resource.ResourceFile) error {
	switch src := s.Controller.Handle.(type) {
	case *configsource.Database:
		err := s.updateDatabase(src, appID, files)
		if err != nil {
			return err
		}
		s.Controller.ReloadApp(appID)
	case *configsource.LocalFS:
		err := s.updateLocalFS(src, appID, files)
		if err != nil {
			return err
		}
		s.Controller.ReloadApp(appID)

	default:
		return errors.New("unsupported configuration source")
	}
	return nil
}

func (s *ConfigService) CreateDomain(appID string, domainID string, domain string, isCustom bool) error {
	if s.DomainImplementation == portalconfig.DomainImplementationTypeKubernetes {
		err := s.Kubernetes.CreateResourcesForDomain(appID, domainID, domain, isCustom)
		if err != nil {
			return fmt.Errorf("failed to create domain k8s resources: %w", err)
		}
	}
	return nil
}

func (s *ConfigService) DeleteDomain(domain *model.Domain) error {
	if s.DomainImplementation == portalconfig.DomainImplementationTypeKubernetes {
		err := s.Kubernetes.DeleteResourcesForDomain(domain.ID)
		if err != nil {
			return fmt.Errorf("failed to delete domain k8s resources: %w", err)
		}
	}
	return nil
}

func (s *ConfigService) updateLocalFS(l *configsource.LocalFS, appID string, updates []*resource.ResourceFile) error {
	fs := l.Fs
	for _, file := range updates {
		if file.Data == nil {
			err := fs.Remove(file.Location.Path)
			// Ignore file not found errors
			if err != nil && !os.IsNotExist(err) {
				return err
			}
		} else {
			err := fs.MkdirAll(filepath.Dir(file.Location.Path), 0777)
			if err != nil {
				return err
			}
			err = afero.WriteFile(fs, file.Location.Path, file.Data, 0666)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *ConfigService) updateDatabase(d *configsource.Database, appID string, updates []*resource.ResourceFile) error {
	return d.UpdateDatabaseSource(appID, updates)
}

func (s *ConfigService) createDatabase(d *configsource.Database, opts *CreateAppOptions) error {
	err := d.CreateDatabaseSource(opts.AppID, opts.Resources, opts.PlanName)
	if err != nil {
		if errors.Is(err, configsource.ErrDuplicatedAppID) {
			return ErrDuplicatedAppID
		}
		return err
	}
	return nil
}
