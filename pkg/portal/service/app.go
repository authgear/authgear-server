package service

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"path"
	"regexp"

	"github.com/spf13/afero"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	globaldb "github.com/authgear/authgear-server/pkg/lib/infra/db/global"

	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/authgear/authgear-server/pkg/portal/model"
	portalresource "github.com/authgear/authgear-server/pkg/portal/resource"
	"github.com/authgear/authgear-server/pkg/portal/util/resources"
	"github.com/authgear/authgear-server/pkg/util/blocklist"
	"github.com/authgear/authgear-server/pkg/util/log"
	corerand "github.com/authgear/authgear-server/pkg/util/rand"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

var ErrAppIDReserved = apierrors.Forbidden.WithReason("AppIDReserved").
	New("requested app ID is reserved")
var ErrAppIDInvalid = apierrors.Invalid.WithReason("InvalidAppID").
	New("invalid app ID")

type AppConfigService interface {
	ResolveContext(appID string) (*config.AppContext, error)
	UpdateResources(appID string, updates []*resource.ResourceFile) error
	Create(opts *CreateAppOptions) error
	CreateDomain(appID string, domainID string, domain string, isCustom bool) error
}

type AppAuthzService interface {
	AddAuthorizedUser(appID string, userID string, role model.CollaboratorRole) error
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
	SQLBuilder  *globaldb.SQLBuilder
	SQLExecutor *globaldb.SQLExecutor

	AppConfig          *portalconfig.AppConfig
	SecretKeyAllowlist portalconfig.SecretKeyAllowlist
	AppConfigs         AppConfigService
	AppAuthz           AppAuthzService
	AppAdminAPI        AppAdminAPIService
	AppDomains         AppDomainService
	Resources          ResourceManager
	AppBaseResources   deps.AppBaseResources
}

func (s *AppService) Get(id string) (*model.App, error) {
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
		app, err := s.Get(id)
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
func (s *AppService) GetMaxOwnedApps(userID string) (int, error) {
	// On errors: ignore and return default quota.

	q := s.SQLBuilder.Select("max_own_apps").
		From(s.SQLBuilder.TableName("_portal_user_app_quota")).
		Where("user_id = ?", userID)
	row, err := s.SQLExecutor.QueryRowWith(q)
	if err != nil {
		return s.AppConfig.MaxOwnedApps, nil
	}

	var quota int
	err = row.Scan(&quota)
	if err != nil {
		return s.AppConfig.MaxOwnedApps, nil
	}

	return quota, nil
}

func (s *AppService) LoadRawAppConfig(app *model.App) (*config.AppConfig, error) {
	result, err := app.Context.Resources.Read(configsource.AppConfig, resource.AppFile{
		Path:              configsource.AuthgearYAML,
		AllowedSecretKeys: s.SecretKeyAllowlist,
	})
	if err != nil {
		return nil, err
	}

	bytes := result.([]byte)
	var cfg *config.AppConfig
	if err := yaml.Unmarshal(bytes, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (s *AppService) LoadRawSecretConfig(app *model.App) (*config.SecretConfig, error) {
	result, err := app.Context.Resources.Read(configsource.SecretConfig, resource.AppFile{
		Path:              configsource.AuthgearSecretYAML,
		AllowedSecretKeys: s.SecretKeyAllowlist,
	})
	if err != nil {
		return nil, err
	}

	bytes := result.([]byte)
	var cfg *config.SecretConfig
	if err := yaml.Unmarshal(bytes, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (s *AppService) Create(userID string, id string) error {
	if err := s.validateAppID(id); err != nil {
		return err
	}

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

	err = s.AppAuthz.AddAuthorizedUser(id, userID, model.CollaboratorRoleOwner)
	if err != nil {
		return err
	}

	return nil
}

func (s *AppService) UpdateResources(app *model.App, updates []resources.Update) error {
	files, err := resources.ApplyUpdates(app.ID, app.Context.AppFs, app.Context.Resources, s.SecretKeyAllowlist, updates)
	if err != nil {
		return err
	}

	err = s.AppConfigs.UpdateResources(app.ID, files)
	return err
}

func (s *AppService) generateResources(appHost string, appID string) (map[string][]byte, error) {
	appResources := make(map[string][]byte)

	// Generate app config
	publicOrigin := &url.URL{Scheme: "https", Host: appHost}
	appConfig := config.GenerateAppConfigFromOptions(&config.GenerateAppConfigOptions{
		AppID:        appID,
		PublicOrigin: publicOrigin.String(),
		CookieDomain: appHost,
	})
	appConfigYAML, err := yaml.Marshal(appConfig)
	if err != nil {
		return nil, err
	}
	appResources[configsource.AuthgearYAML] = appConfigYAML

	// Generate secret config
	secretConfig := config.GenerateSecretConfigFromOptions(&config.GenerateSecretConfigOptions{}, corerand.SecureRand)
	secretConfigYAML, err := yaml.Marshal(secretConfig)
	if err != nil {
		return nil, err
	}
	appResources[configsource.AuthgearSecretYAML] = secretConfigYAML

	return appResources, nil
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
		err = ErrAppIDInvalid
		return
	}

	files, err := s.generateResources(appHost, appID)
	if err != nil {
		return
	}

	fs := afero.NewMemMapFs()
	for p, data := range files {
		_ = fs.MkdirAll(path.Dir(p), 0777)
		_ = afero.WriteFile(fs, p, data, 0666)
	}

	appFs := resource.LeveledAferoFs{Fs: fs, FsLevel: resource.FsLevelApp}
	resMgr := (*resource.Manager)(s.AppBaseResources).Overlay(appFs)
	_, err = resources.ApplyUpdates(appID, appFs, resMgr, s.SecretKeyAllowlist, nil)
	if err != nil {
		return
	}

	opts = &CreateAppOptions{
		AppID:     appID,
		Resources: files,
	}

	return
}

func (s *AppService) validateAppID(appID string) error {
	var list *blocklist.Blocklist
	result, err := s.Resources.Read(portalresource.ReservedAppIDTXT, resource.EffectiveResource{})
	if errors.Is(err, resource.ErrResourceNotFound) {
		// No reserved usernames
		list = &blocklist.Blocklist{}
	} else if err != nil {
		return err
	} else {
		list = result.(*blocklist.Blocklist)
	}

	if list.IsBlocked(appID) {
		return ErrAppIDReserved
	}

	return nil
}
