package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"
	texttemplate "text/template"

	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/log"
	corerand "github.com/authgear/authgear-server/pkg/util/rand"
)

type AppConfigService interface {
	ResolveContext(appID string) (*config.AppContext, error)
	UpdateConfig(appID string, updateFiles []*model.AppConfigFile, deleteFiles []string) error
	Create(id string, appConfigYAML []byte, secretConfigYAML []byte) error
}

type AppAuthzService interface {
	AddAuthorizedUser(appID string, userID string) error
	ListAuthorizedApps(userID string) ([]string, error)
}

type AppServiceLogger struct{ *log.Logger }

func NewAppServiceLogger(lf *log.Factory) AppServiceLogger {
	return AppServiceLogger{lf.New("app-service")}
}

type AppService struct {
	Logger     AppServiceLogger
	AppConfig  *portalconfig.AppConfig
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

func (s *AppService) Create(userID string, id string) error {
	s.Logger.
		WithField("user_id", userID).
		WithField("app_id", id).
		Info("creating app")

	appConfigYAML, secretConfigYAML, err := s.generateConfig(id)
	if err != nil {
		return err
	}

	err = s.AppConfigs.Create(id, appConfigYAML, secretConfigYAML)
	if err != nil {
		return err
	}

	err = s.AppAuthz.AddAuthorizedUser(id, userID)
	if err != nil {
		return err
	}

	return nil
}

func (s *AppService) UpdateConfig(app *model.App, updateFiles []*model.AppConfigFile, deleteFiles []string) error {
	err := ValidateConfig(app.ID, *app.Context.Config, updateFiles, deleteFiles)
	if err != nil {
		return err
	}
	err = s.AppConfigs.UpdateConfig(app.ID, updateFiles, deleteFiles)
	return err
}

func (s *AppService) generateAppConfig(appID string) (*config.AppConfig, error) {
	if s.AppConfig.HostTemplate == "" {
		return nil, errors.New("app hostname template is not configured")
	}
	t := texttemplate.New("host-template")
	_, err := t.Parse(s.AppConfig.HostTemplate)
	if err != nil {
		return nil, err
	}
	var buf strings.Builder

	data := map[string]interface{}{
		"AppID": appID,
	}
	err = t.Execute(&buf, data)
	if err != nil {
		return nil, err
	}

	appOrigin := &url.URL{Scheme: "https", Host: buf.String()}
	cfg := config.GenerateAppConfigFromOptions(&config.GenerateAppConfigOptions{
		AppID:        appID,
		PublicOrigin: appOrigin.String(),
	})
	return cfg, nil
}

func (s *AppService) generateSecretConfig() (*config.SecretConfig, error) {
	secrets := s.AppConfig.Secret
	cfg := config.GenerateSecretConfigFromOptions(&config.GenerateSecretConfigOptions{
		DatabaseURL:    secrets.DatabaseURL,
		DatabaseSchema: secrets.DatabaseSchema,
		RedisURL:       secrets.RedisURL,
	}, corerand.SecureRand)

	if secrets.SMTPHost != "" {
		data, err := json.Marshal(&config.SMTPServerCredentials{
			Host:     secrets.SMTPHost,
			Port:     secrets.SMTPPort,
			Mode:     config.SMTPMode(secrets.SMTPMode),
			Username: secrets.SMTPUsername,
			Password: secrets.SMTPPassword,
		})
		if err != nil {
			panic(err)
		}

		cfg.Secrets = append(cfg.Secrets, config.SecretItem{
			Key:     config.SMTPServerCredentialsKey,
			RawData: data,
		})
	}

	if secrets.TwilioAccountSID != "" {
		data, err := json.Marshal(&config.TwilioCredentials{
			AccountSID: secrets.TwilioAccountSID,
			AuthToken:  secrets.TwilioAuthToken,
		})
		if err != nil {
			panic(err)
		}

		cfg.Secrets = append(cfg.Secrets, config.SecretItem{
			Key:     config.TwilioCredentialsKey,
			RawData: data,
		})
	}

	if secrets.NexmoAPIKey != "" {
		data, err := json.Marshal(&config.NexmoCredentials{
			APIKey:    secrets.NexmoAPIKey,
			APISecret: secrets.NexmoAPISecret,
		})
		if err != nil {
			panic(err)
		}

		cfg.Secrets = append(cfg.Secrets, config.SecretItem{
			Key:     config.NexmoCredentialsKey,
			RawData: data,
		})
	}

	return cfg, nil
}

func (s *AppService) generateConfig(appID string) (appConfigYAML []byte, secretConfigYAML []byte, err error) {
	appIDRegex, err := regexp.Compile(s.AppConfig.IDPattern)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid app ID validation pattern: %w", err)
	}
	if !appIDRegex.MatchString(appID) {
		return nil, nil, apierrors.NewInvalid("invalid app ID")
	}

	appConfig, err := s.generateAppConfig(appID)
	if err != nil {
		return nil, nil, err
	}
	appConfigYAML, err = yaml.Marshal(appConfig)
	if err != nil {
		return nil, nil, err
	}

	secretConfig, err := s.generateSecretConfig()
	if err != nil {
		return nil, nil, err
	}
	secretConfigYAML, err = yaml.Marshal(secretConfig)
	if err != nil {
		return nil, nil, err
	}

	err = ValidateConfig(appID, config.Config{}, []*model.AppConfigFile{
		{Path: configsource.AuthgearYAML, Content: string(appConfigYAML)},
		{Path: configsource.AuthgearSecretYAML, Content: string(secretConfigYAML)},
	}, nil)
	if err != nil {
		return nil, nil, err
	}

	return appConfigYAML, secretConfigYAML, nil
}

const ConfigFileMaxSize = 100 * 1024

func ValidateConfig(appID string, cfg config.Config, updateFiles []*model.AppConfigFile, deleteFiles []string) error {
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
	for _, file := range updateFiles {
		if file.Path == "/"+configsource.AuthgearYAML {
			appConfig, err := config.Parse([]byte(file.Content))
			if err != nil {
				return fmt.Errorf("%s is invalid: %w", file.Path, err)
			} else if string(appConfig.ID) != appID {
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
