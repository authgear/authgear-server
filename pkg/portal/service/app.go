package service

import (
	"errors"
	"fmt"
	"path"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/portal/model"
)

type AppAuthzService interface {
	ListAuthorizedApps(userID string) ([]string, error)
}

type AppService struct {
	ConfigSource *configsource.ConfigSource
	AppAuthz     AppAuthzService
}

func (s *AppService) GetMany(ids []string) (out []*model.App, err error) {
	for _, id := range ids {
		appCtx, err := s.ConfigSource.ContextResolver.ResolveContext(id)
		if err != nil {
			return nil, err
		}
		out = append(out, &model.App{
			ID:      id,
			Context: appCtx,
		})
	}

	return
}

func (s *AppService) List(userID string) ([]*model.App, error) {
	appIDs, err := s.AppAuthz.ListAuthorizedApps(userID)
	if err != nil {
		return nil, err
	}

	return s.GetMany(appIDs)
}

func (s *AppService) UpdateConfig(app *model.App, updateFiles []*model.AppConfigFile, deleteFiles []string) error {
	err := ValidateConfig(app, updateFiles, deleteFiles)
	if err != nil {
		return err
	}
	// TODO(portal): update files
	fmt.Printf("%v %#v %#v\n", app, updateFiles, deleteFiles)
	return errors.New("??e")
}

const ConfigFileMaxSize = 100 * 1024

func ValidateConfig(app *model.App, updateFiles []*model.AppConfigFile, deleteFiles []string) error {
	// Normalize the paths.
	for _, file := range updateFiles {
		file.Path = path.Clean("/" + file.Path)
	}
	for i, p := range deleteFiles {
		deleteFiles[i] = path.Clean("/" + p)
	}

	// Forbid deleting configuration YAML.
	for _, p := range deleteFiles {
		if p == "/"+configsource.AuthgearYAML || p == "/"+configsource.AuthgearSecretYAML {
			return errors.New("cannot delete main configuration YAML files")
		}
	}

	// Validate file size.
	for _, f := range updateFiles {
		if len(f.Content) > ConfigFileMaxSize {
			return fmt.Errorf("%s is too large: %v > %v", f.Path, len(f.Content), ConfigFileMaxSize)
		}
	}

	// Validate configuration YAML.
	cfg := *app.Context.Config
	for _, file := range updateFiles {
		if file.Path == "/"+configsource.AuthgearYAML {
			appConfig, err := config.Parse([]byte(file.Content))
			if err != nil {
				return fmt.Errorf("%s is invalid: %w", file.Path, err)
			} else if string(appConfig.ID) != app.ID {
				return fmt.Errorf("%s is invalid: invalid app ID", file.Path)
			}
			cfg.AppConfig = appConfig
		} else if file.Path == "/"+configsource.AuthgearSecretYAML {
			secretConfig, err := config.ParseSecret([]byte(file.Content))
			if err != nil {
				return fmt.Errorf("%s is invalid: %w", file.Path, err)
			}
			cfg.SecretConfig = secretConfig
		}
	}
	err := cfg.SecretConfig.Validate(cfg.AppConfig)
	if err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// TODO(portal): validate templates.

	return nil
}
