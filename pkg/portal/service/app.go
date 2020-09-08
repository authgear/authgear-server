package service

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/portal/model"
)

type AppConfigService interface {
	ResolveContext(appID string) (*config.AppContext, error)
	UpdateConfig(appID string, updateFiles []*model.AppConfigFile, deleteFiles []string) error
}

type AppAuthzService interface {
	ListAuthorizedApps(userID string) ([]string, error)
}

type AppService struct {
	AppConfigs AppConfigService
	AppAuthz   AppAuthzService
}

func (s *AppService) GetMany(ids []string) (out []*model.App, err error) {
	for _, id := range ids {
		appCtx, err := s.AppConfigs.ResolveContext(id)
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
	err = s.AppConfigs.UpdateConfig(app.ID, updateFiles, deleteFiles)
	return err
}

const ConfigFileMaxSize = 100 * 1024

func ValidateConfig(app *model.App, updateFiles []*model.AppConfigFile, deleteFiles []string) error {
	// Validate the file names.
	for _, file := range updateFiles {
		if filepath.Base(file.Name) != file.Name {
			return fmt.Errorf("invalid file name: %s", file.Name)
		}
	}
	for _, fileName := range deleteFiles {
		if filepath.Base(fileName) != fileName {
			return fmt.Errorf("invalid file name: %s", fileName)
		}
	}

	// Forbid deleting configuration YAML.
	for _, name := range deleteFiles {
		if name == configsource.AuthgearYAML || name == configsource.AuthgearSecretYAML {
			return errors.New("cannot delete main configuration YAML files")
		}
	}

	// Validate file size.
	for _, f := range updateFiles {
		if len(f.Content) > ConfigFileMaxSize {
			return fmt.Errorf("%s is too large: %v > %v", f.Name, len(f.Content), ConfigFileMaxSize)
		}
	}

	// Validate configuration YAML.
	cfg := *app.Context.Config
	for _, file := range updateFiles {
		if file.Name == configsource.AuthgearYAML {
			appConfig, err := config.Parse([]byte(file.Content))
			if err != nil {
				return fmt.Errorf("%s is invalid: %w", file.Name, err)
			} else if string(appConfig.ID) != app.ID {
				return fmt.Errorf("%s is invalid: invalid app ID", file.Name)
			}
			cfg.AppConfig = appConfig
		} else if file.Name == configsource.AuthgearSecretYAML {
			secretConfig, err := config.ParseSecret([]byte(file.Content))
			if err != nil {
				return fmt.Errorf("%s is invalid: %w", file.Name, err)
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
