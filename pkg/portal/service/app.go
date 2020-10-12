package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"path"
	"regexp"
	"strings"
	texttemplate "text/template"

	"github.com/spf13/afero"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/translation"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/fs"
	"github.com/authgear/authgear-server/pkg/util/log"
	corerand "github.com/authgear/authgear-server/pkg/util/rand"
)

const RedactedValue = "<REDACTED>"

type redactionMapping struct {
	target *string
	secret string
}

type generateAppConfigAndTranslationResult struct {
	AppConfig           *config.AppConfig
	TranslationJSONPath string
	TranslationJSON     []byte
}

type AppConfigService interface {
	ResolveContext(appID string) (*config.AppContext, error)
	UpdateConfig(appID string, updateFiles []*model.AppConfigFile, deleteFiles []string) error
	Create(opts *CreateAppOptions) error
	CreateDomain(appID string, domain *model.Domain) error
}

type AppAuthzService interface {
	AddAuthorizedUser(appID string, userID string) error
	ListAuthorizedApps(userID string) ([]string, error)
}

type AppAdminAPIService interface {
	ResolveHost(appID string) (host string, err error)
}

type AppDomainService interface {
	CreateDomain(appID string, domain string, isVerified bool, isCustom bool) (*model.Domain, error)
}

type AppServiceLogger struct{ *log.Logger }

func NewAppServiceLogger(lf *log.Factory) AppServiceLogger {
	return AppServiceLogger{lf.New("app-service")}
}

type AppService struct {
	Logger      AppServiceLogger
	AppConfig   *portalconfig.AppConfig
	AppConfigs  AppConfigService
	AppAuthz    AppAuthzService
	AppAdminAPI AppAdminAPIService
	AppDomains  AppDomainService
}

func (s *AppService) loadApp(id string) (*model.App, error) {
	appCtx, err := s.AppConfigs.ResolveContext(id)
	if err != nil {
		return nil, err
	}

	secretConfig := *appCtx.Config.SecretConfig
	err = s.redactSecrets(&secretConfig)
	if err != nil {
		return nil, err
	}

	secretConfigYAML, err := yaml.Marshal(secretConfig)
	if err != nil {
		return nil, err
	}

	overlayFs := afero.NewBasePathFs(afero.NewMemMapFs(), "/")
	err = afero.WriteFile(overlayFs, configsource.AuthgearSecretYAML, secretConfigYAML, 0666)
	if err != nil {
		return nil, err
	}

	appCtx = &config.AppContext{
		Fs: &fs.OverlayFs{Base: appCtx.Fs, Overlay: &fs.AferoFs{Fs: overlayFs}},
		Config: &config.Config{
			AppConfig:    appCtx.Config.AppConfig,
			SecretConfig: &secretConfig,
		},
	}

	return &model.App{
		ID:      id,
		Context: appCtx,
	}, nil
}

func (s *AppService) GetMany(ids []string) (out []*model.App, err error) {
	for _, id := range ids {
		app, err := s.loadApp(id)
		if err != nil {
			return nil, err
		}
		out = append(out, app)
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

	appHost, err := s.generateAppHost(id)
	if err != nil {
		return err
	}

	createAppOpts, err := s.generateConfig(appHost, id)
	if err != nil {
		return err
	}

	adminAPIHost, err := s.AppAdminAPI.ResolveHost(id)
	if err != nil {
		return err
	}

	err = s.AppConfigs.Create(createAppOpts)
	if err != nil {
		// TODO(portal): cleanup orphaned resources created from failed app creation
		s.Logger.WithError(err).WithField("app_id", id).Error("failed to create app")
		return err
	}

	// Deduplicate hosts
	hosts := map[string]struct{}{
		appHost:      {},
		adminAPIHost: {},
	}
	for host := range hosts {
		domain := host
		if h, _, err := net.SplitHostPort(host); err == nil {
			domain = h
		}

		_, err := s.AppDomains.CreateDomain(id, domain, true, false)
		if err != nil {
			return err
		}
	}

	err = s.AppAuthz.AddAuthorizedUser(id, userID)
	if err != nil {
		return err
	}

	return nil
}

func (s *AppService) UpdateConfig(app *model.App, updateFiles []*model.AppConfigFile, deleteFiles []string) error {
	PrepareUpdates(updateFiles, deleteFiles)

	// Un-redact secrets
	for _, file := range updateFiles {
		if file.Path == "/"+configsource.AuthgearSecretYAML {
			secretConfig, err := config.ParseSecret([]byte(file.Content))
			if err != nil {
				return err
			}
			if err := s.unredactSecrets(secretConfig); err != nil {
				return err
			}
			cfgYAML, err := yaml.Marshal(secretConfig)
			if err != nil {
				return err
			}
			file.Content = string(cfgYAML)
		}
	}

	err := ValidateConfig(app.ID, *app.Context.Config, updateFiles, deleteFiles)
	if err != nil {
		return err
	}

	err = s.AppConfigs.UpdateConfig(app.ID, updateFiles, deleteFiles)
	return err
}

func (s *AppService) generateAppConfigAndTranslation(appHost string, appID string) (*generateAppConfigAndTranslationResult, error) {
	appOrigin := &url.URL{Scheme: "https", Host: appHost}
	cfg := config.GenerateAppConfigFromOptions(&config.GenerateAppConfigOptions{
		AppID:        appID,
		PublicOrigin: appOrigin.String(),
	})

	translationObj := make(map[string]string)
	if s.AppConfig.Branding.AppName != "" {
		translationObj["app.app-name"] = s.AppConfig.Branding.AppName
	}
	if s.AppConfig.Branding.EmailDefaultSender != "" {
		translationObj["email.default.sender"] = s.AppConfig.Branding.EmailDefaultSender
	}
	if s.AppConfig.Branding.SMSDefaultSender != "" {
		translationObj["sms.default.sender"] = s.AppConfig.Branding.SMSDefaultSender
	}

	var err error
	var translationJSON []byte
	var translationJSONPath string
	if len(translationObj) > 0 {
		translationJSONPath = "templates/" + translation.TemplateItemTypeTranslationJSON
		translationJSON, err = json.Marshal(translationObj)
		if err != nil {
			return nil, err
		}
		cfg.Template = &config.TemplateConfig{
			Items: []config.TemplateItem{
				config.TemplateItem{
					Type: config.TemplateItemType(translation.TemplateItemTypeTranslationJSON),
					URI:  "file:///" + translationJSONPath,
				},
			},
		}
	}

	result := &generateAppConfigAndTranslationResult{
		AppConfig:           cfg,
		TranslationJSONPath: translationJSONPath,
		TranslationJSON:     translationJSON,
	}
	return result, nil
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

func (s *AppService) generateAppHost(appID string) (string, error) {
	if s.AppConfig.HostTemplate == "" {
		return "", errors.New("app hostname template is not configured")
	}
	t := texttemplate.New("host-template")
	_, err := t.Parse(s.AppConfig.HostTemplate)
	if err != nil {
		return "", err
	}
	var buf strings.Builder

	data := map[string]interface{}{
		"AppID": appID,
	}
	err = t.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (s *AppService) generateConfig(appHost string, appID string) (opts *CreateAppOptions, err error) {
	appIDRegex, err := regexp.Compile(s.AppConfig.IDPattern)
	if err != nil {
		err = fmt.Errorf("invalid app ID validation pattern: %w", err)
		return
	}
	if !appIDRegex.MatchString(appID) {
		err = apierrors.NewInvalid("invalid app ID")
		return
	}

	genResult, err := s.generateAppConfigAndTranslation(appHost, appID)
	if err != nil {
		return
	}
	appConfigYAML, err := yaml.Marshal(genResult.AppConfig)
	if err != nil {
		return
	}

	secretConfig, err := s.generateSecretConfig()
	if err != nil {
		return
	}
	secretConfigYAML, err := yaml.Marshal(secretConfig)
	if err != nil {
		return
	}

	err = ValidateConfig(appID, config.Config{}, []*model.AppConfigFile{
		{Path: "/" + configsource.AuthgearYAML, Content: string(appConfigYAML)},
		{Path: "/" + configsource.AuthgearSecretYAML, Content: string(secretConfigYAML)},
	}, nil)
	if err != nil {
		return
	}

	opts = &CreateAppOptions{
		AppID:               appID,
		AppConfigYAML:       appConfigYAML,
		SecretConfigYAML:    secretConfigYAML,
		TranslationJSONPath: genResult.TranslationJSONPath,
		TranslationJSON:     genResult.TranslationJSON,
	}

	return
}

func (s *AppService) redactionMappings(cfg *config.SecretConfig) []redactionMapping {
	var mappings []redactionMapping
	addMapping := func(key config.SecretKey, emptyData config.SecretItemData, secret string, mapFn func(s config.SecretItemData) *string) {
		// If the cluster secret is undefined, do not add the mapping.
		if secret == "" {
			return
		}

		// Otherwise redaction always occur even if the secret item has been removed from the secret config.
		// This means if the secret config has the secret item missing,
		// after redaction the item would be added again.
		item, ok := cfg.Lookup(key)
		if !ok {
			cfg.Secrets = append(cfg.Secrets, config.SecretItem{
				Key:  key,
				Data: emptyData,
			})
			item = &cfg.Secrets[len(cfg.Secrets)-1]
		}

		mappings = append(mappings, redactionMapping{
			target: mapFn(item.Data),
			secret: secret,
		})
	}

	addMapping(
		config.DatabaseCredentialsKey,
		&config.DatabaseCredentials{},
		s.AppConfig.Secret.DatabaseURL,
		func(s config.SecretItemData) *string {
			return &s.(*config.DatabaseCredentials).DatabaseURL
		},
	)
	addMapping(
		config.DatabaseCredentialsKey,
		&config.DatabaseCredentials{},
		s.AppConfig.Secret.DatabaseSchema,
		func(s config.SecretItemData) *string {
			return &s.(*config.DatabaseCredentials).DatabaseSchema
		},
	)
	addMapping(
		config.RedisCredentialsKey,
		&config.RedisCredentials{},
		s.AppConfig.Secret.RedisURL,
		func(s config.SecretItemData) *string {
			return &s.(*config.RedisCredentials).RedisURL
		},
	)
	addMapping(
		config.SMTPServerCredentialsKey,
		&config.SMTPServerCredentials{},
		s.AppConfig.Secret.SMTPHost,
		func(s config.SecretItemData) *string {
			return &s.(*config.SMTPServerCredentials).Host
		},
	)
	addMapping(
		config.SMTPServerCredentialsKey,
		&config.SMTPServerCredentials{},
		s.AppConfig.Secret.SMTPUsername,
		func(s config.SecretItemData) *string {
			return &s.(*config.SMTPServerCredentials).Username
		},
	)
	addMapping(
		config.SMTPServerCredentialsKey,
		&config.SMTPServerCredentials{},
		s.AppConfig.Secret.SMTPPassword,
		func(s config.SecretItemData) *string {
			return &s.(*config.SMTPServerCredentials).Password
		},
	)
	addMapping(
		config.TwilioCredentialsKey,
		&config.TwilioCredentials{},
		s.AppConfig.Secret.TwilioAccountSID,
		func(s config.SecretItemData) *string {
			return &s.(*config.TwilioCredentials).AccountSID
		},
	)
	addMapping(
		config.TwilioCredentialsKey,
		&config.TwilioCredentials{},
		s.AppConfig.Secret.TwilioAuthToken,
		func(s config.SecretItemData) *string {
			return &s.(*config.TwilioCredentials).AuthToken
		},
	)
	addMapping(
		config.NexmoCredentialsKey,
		&config.NexmoCredentials{},
		s.AppConfig.Secret.NexmoAPIKey,
		func(s config.SecretItemData) *string {
			return &s.(*config.NexmoCredentials).APIKey
		},
	)
	addMapping(
		config.NexmoCredentialsKey,
		&config.NexmoCredentials{},
		s.AppConfig.Secret.NexmoAPISecret,
		func(s config.SecretItemData) *string {
			return &s.(*config.NexmoCredentials).APISecret
		},
	)

	return mappings
}

func (s *AppService) redactSecrets(cfg *config.SecretConfig) error {
	for _, mapping := range s.redactionMappings(cfg) {
		// Add back the value as redacted even if the item is missing
		if *mapping.target == "" || *mapping.target == mapping.secret {
			*mapping.target = RedactedValue
		}
	}
	for i, item := range cfg.Secrets {
		data, err := json.Marshal(item.Data)
		if err != nil {
			return err
		}

		item.RawData = data
		cfg.Secrets[i] = item
	}
	return nil
}

func (s *AppService) unredactSecrets(cfg *config.SecretConfig) error {
	for _, mapping := range s.redactionMappings(cfg) {
		// TODO(portal): allow bring-in credentials
		if *mapping.target != RedactedValue {
			return apierrors.NewForbidden("cannot change secret value")
		}
		*mapping.target = mapping.secret
	}
	for i, item := range cfg.Secrets {
		data, err := json.Marshal(item.Data)
		if err != nil {
			return err
		}

		item.RawData = data
		cfg.Secrets[i] = item
	}
	return nil
}

const ConfigFileMaxSize = 100 * 1024

func PrepareUpdates(updateFiles []*model.AppConfigFile, deleteFiles []string) {
	// Normalize the paths.
	for _, file := range updateFiles {
		file.Path = path.Clean("/" + file.Path)
	}
	for i, p := range deleteFiles {
		deleteFiles[i] = path.Clean("/" + p)
	}
}

func ValidateConfig(appID string, cfg config.Config, updateFiles []*model.AppConfigFile, deleteFiles []string) error {
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

	templatePaths := map[string]struct{}{}
	oldTemplatePaths := map[string]struct{}{}
	for _, item := range cfg.AppConfig.Template.Items {
		u, err := url.Parse(item.URI)
		if err != nil {
			return fmt.Errorf("invalid URI for template '%s': %w", item.Type, err)
		}
		if u.Scheme != "file" {
			return fmt.Errorf("invalid URI for template '%s': only 'file' scheme is supported", item.Type)
		}
		if u.Path != path.Clean(u.Path) {
			return fmt.Errorf("invalid URI for template '%s': path must be normalized", item.Type)
		}
		templatePaths[u.Path] = struct{}{}
	}
	nullableOldConfig := cfg.AppConfig
	if nullableOldConfig != nil {
		for _, item := range nullableOldConfig.Template.Items {
			u, err := url.Parse(item.URI)
			if err != nil || u.Scheme != "file" {
				continue
			}
			oldTemplatePaths[u.Path] = struct{}{}
		}
	}

	for _, f := range updateFiles {
		if f.Path == "/"+configsource.AuthgearYAML || f.Path == "/"+configsource.AuthgearSecretYAML {
			continue
		}
		if _, ok := templatePaths[f.Path]; !ok {
			return fmt.Errorf("invalid file '%s': file is not referenced from configuration", f.Path)
		}
	}
	for _, p := range deleteFiles {
		// Forbid deleting configuration YAML.
		if p == "/"+configsource.AuthgearYAML || p == "/"+configsource.AuthgearSecretYAML {
			return errors.New("cannot delete main configuration YAML files")
		}
		if _, ok := oldTemplatePaths[p]; !ok {
			return fmt.Errorf("invalid file '%s': file is not referenced from configuration", p)
		}
	}

	return nil
}
