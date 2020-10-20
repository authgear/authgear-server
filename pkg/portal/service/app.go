package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp"

	"github.com/spf13/afero"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/portal/resources"
	"github.com/authgear/authgear-server/pkg/util/log"
	corerand "github.com/authgear/authgear-server/pkg/util/rand"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/template"
)

type generateAppConfigAndTranslationResult struct {
	AppConfig           *config.AppConfig
	TranslationJSONPath string
	TranslationJSON     []byte
}

type AppConfigService interface {
	ResolveContext(appID string) (*config.AppContext, error)
	UpdateResources(appID string, updates []resources.Update) error
	Create(opts *CreateAppOptions) error
	CreateDomain(appID string, domainID string, domain string, isCustom bool) error
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

type AppBaseResource *resource.Manager

type AppService struct {
	Logger           AppServiceLogger
	AppConfig        *portalconfig.AppConfig
	AppConfigs       AppConfigService
	AppAuthz         AppAuthzService
	AppAdminAPI      AppAdminAPIService
	AppDomains       AppDomainService
	AppBaseResources deps.AppBaseResources
}

func (s *AppService) loadApp(id string) (*model.App, error) {
	appCtx, err := s.AppConfigs.ResolveContext(id)
	if err != nil {
		return nil, err
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
		isMain := host == appHost

		domain := host
		if h, _, err := net.SplitHostPort(host); err == nil {
			domain = h
		}

		if isMain {
			_, err := s.AppDomains.CreateDomain(id, domain, true, false)
			if err != nil {
				return err
			}
		} else {
			err := s.AppConfigs.CreateDomain(id, "", domain, false)
			if err != nil {
				return err
			}
		}
	}

	err = s.AppAuthz.AddAuthorizedUser(id, userID)
	if err != nil {
		return err
	}

	return nil
}

func (s *AppService) UpdateResources(app *model.App, updates []resources.Update) error {
	err := resources.Validate(app.ID, app.Context.AppFs, app.Context.Resources, updates)
	if err != nil {
		return err
	}

	err = s.AppConfigs.UpdateResources(app.ID, updates)
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
		// FIXME(resource): This fix is temporary.
		translationJSONPath = "templates/__default__/" + template.TranslationJSONName
		translationJSON, err = json.Marshal(translationObj)
		if err != nil {
			return nil, err
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
	if s.AppConfig.HostSuffix == "" {
		return "", errors.New("app hostname suffix is not configured")
	}
	return appID + s.AppConfig.HostSuffix, nil
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

	// FIXME(resource): allow providing resource FS template for new apps
	appFs := resource.AferoFs{Fs: afero.NewMemMapFs()}

	resMgr := (*resource.Manager)(s.AppBaseResources).Overlay(appFs)
	updates := []resources.Update{
		{Path: configsource.AuthgearYAML, Data: appConfigYAML},
		{Path: configsource.AuthgearSecretYAML, Data: secretConfigYAML},
	}
	err = resources.Validate(appID, appFs, resMgr, updates)
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
