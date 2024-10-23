package service

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"time"

	"github.com/lib/pq"
	"github.com/spf13/afero"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/lib/saml"
	"github.com/authgear/authgear-server/pkg/lib/tester"
	"github.com/authgear/authgear-server/pkg/portal/appresource"
	"github.com/authgear/authgear-server/pkg/portal/appsecret"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/portal/model"
	portalresource "github.com/authgear/authgear-server/pkg/portal/resource"
	"github.com/authgear/authgear-server/pkg/util/blocklist"
	"github.com/authgear/authgear-server/pkg/util/checksum"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/intl"
	"github.com/authgear/authgear-server/pkg/util/log"
	corerand "github.com/authgear/authgear-server/pkg/util/rand"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/template"
)

const DefaultTermsOfServiceLink string = "https://www.authgear.com/terms"
const DefaultPrivacyPolicyLink string = "https://www.authgear.com/data-privacy"
const SecretVisitTokenValidDuration = duration.Short
const SecretVisitTokenVisibleSecrets string = "visible_secrets"

var ErrAppIDReserved = apierrors.Forbidden.WithReason("AppIDReserved").
	New("requested app ID is reserved")
var ErrAppIDInvalid = apierrors.Invalid.WithReason("InvalidAppID").
	New("invalid app ID")
var ErrReauthRequrired = apierrors.Forbidden.WithReason("ReauthRequrired").
	New("reauthentication required")

type AppConfigService interface {
	ResolveContext(appID string) (*config.AppContext, error)
	UpdateResources(appID string, updates []*resource.ResourceFile) error
	Create(opts *CreateAppOptions) error
}

type AppAuthzService interface {
	AddAuthorizedUser(appID string, userID string, role model.CollaboratorRole) error
	ListAuthorizedApps(userID string) ([]string, error)
}

type AppDefaultDomainService interface {
	GetLatestAppHost(appID string) (string, error)
	CreateAllDefaultDomains(appID string) error
}

type AppPlanService interface {
	GetDefaultPlan() (*model.Plan, error)
}

type AppServiceLogger struct{ *log.Logger }

func NewAppServiceLogger(lf *log.Factory) AppServiceLogger {
	return AppServiceLogger{lf.New("app-service")}
}

type AppResourceManagerFactory interface {
	NewManagerWithNewAppFS(appFs resource.Fs) *appresource.Manager
	NewManagerWithAppContext(appContext *config.AppContext) *appresource.Manager
}

type AppSecretVisitTokenStore interface {
	CreateToken(
		appID config.AppID,
		userID string,
		secrets []config.SecretKey,
	) (*appsecret.AppSecretVisitToken, error)
	GetTokenByID(
		appID config.AppID,
		tokenID string,
	) (*appsecret.AppSecretVisitToken, error)
}

type AppTesterTokenStore interface {
	CreateToken(
		appID config.AppID,
		returnURI string,
	) (*tester.TesterToken, error)
}

type AppService struct {
	Logger      AppServiceLogger
	SQLBuilder  *globaldb.SQLBuilder
	SQLExecutor *globaldb.SQLExecutor

	GlobalDatabase *globaldb.Handle

	AppConfig                *portalconfig.AppConfig
	AppConfigs               AppConfigService
	AppAuthz                 AppAuthzService
	DefaultDomains           AppDefaultDomainService
	Resources                ResourceManager
	AppResMgrFactory         AppResourceManagerFactory
	Plan                     AppPlanService
	Clock                    clock.Clock
	AppSecretVisitTokenStore AppSecretVisitTokenStore
	AppTesterTokenStore      AppTesterTokenStore
	SAMLEnvironmentConfig    config.SAMLEnvironmentConfig
}

// Get calls other services that acquires connection themselves.
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

// GetMany just uses Get.
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

// GetAppList calls other services that acquires connection themselves.
func (s *AppService) GetAppList(userID string) ([]*model.AppListItem, error) {
	appIDs, err := s.AppAuthz.ListAuthorizedApps(userID)
	if err != nil {
		return nil, err
	}

	apps, err := s.GetMany(appIDs)
	if err != nil {
		return nil, err
	}

	appList := []*model.AppListItem{}
	for _, app := range apps {
		appList = append(appList, &model.AppListItem{
			AppID:        app.ID,
			PublicOrigin: app.Context.Config.AppConfig.HTTP.PublicOrigin,
		})
	}
	return appList, nil
}

// GetProjectQuota acquires connection.
func (s *AppService) GetProjectQuota(userID string) (int, error) {
	q := s.SQLBuilder.Select("max_own_apps").
		From(s.SQLBuilder.TableName("_portal_user_app_quota")).
		Where("user_id = ?", userID)

	var quota int
	err := s.GlobalDatabase.WithTx(func() error {
		row, err := s.SQLExecutor.QueryRowWith(q)
		if err != nil {
			return err
		}
		err = row.Scan(&quota)
		// Use the default quota if this user has no specific quota.
		if errors.Is(err, sql.ErrNoRows) {
			quota = s.AppConfig.MaxOwnedApps
			return nil
		} else if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return quota, nil
}

// GetManyProjectQuota acquires connection.
func (s *AppService) GetManyProjectQuota(userIDs []string) ([]int, error) {
	q := s.SQLBuilder.Select(
		"user_id",
		"max_own_apps",
	).
		From(s.SQLBuilder.TableName("_portal_user_app_quota")).
		Where("user_id = ANY (?)", pq.Array(userIDs))

	m := make(map[string]int)
	err := s.GlobalDatabase.WithTx(func() error {
		rows, err := s.SQLExecutor.QueryWith(q)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var userID string
			var count int
			err = rows.Scan(&userID, &count)
			if err != nil {
				return err
			}
			m[userID] = count
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	out := make([]int, len(userIDs))
	for i, userID := range userIDs {
		if count, ok := m[userID]; ok {
			out[i] = count
		} else {
			out[i] = s.AppConfig.MaxOwnedApps
		}
	}

	return out, nil
}

// LoadRawAppConfig does not need connection.
func (s *AppService) LoadRawAppConfig(app *model.App) (*config.AppConfig, string, error) {
	resMgr := s.AppResMgrFactory.NewManagerWithAppContext(app.Context)
	result, err := resMgr.ReadAppFile(configsource.AppConfig,
		&resource.AppFile{
			Path: configsource.AuthgearYAML,
		})
	if err != nil {
		return nil, "", err
	}

	bytes := result.([]byte)
	checksum := checksum.CRC32IEEEInHex(bytes)
	var cfg *config.AppConfig
	if err := yaml.Unmarshal(bytes, &cfg); err != nil {
		return nil, "", err
	}
	return cfg, checksum, nil
}

// LoadAppSecretConfig does not need connection.
func (s *AppService) LoadAppSecretConfig(
	app *model.App,
	sessionInfo *apimodel.SessionInfo,
	token string) (*model.SecretConfig, string, error) {
	var unmaskedSecrets []config.SecretKey = []config.SecretKey{}
	resMgr := s.AppResMgrFactory.NewManagerWithAppContext(app.Context)
	result, err := resMgr.ReadAppFile(configsource.SecretConfig, &resource.AppFile{
		Path: configsource.AuthgearSecretYAML,
	})
	if err != nil {
		return nil, "", err
	}

	bytes := result.([]byte)
	checksum := checksum.CRC32IEEEInHex(bytes)

	cfg, err := config.ParsePartialSecret(bytes)
	if err != nil {
		return nil, "", err
	}

	now := s.Clock.NowUTC()
	if token != "" {
		tokenModel, err := s.AppSecretVisitTokenStore.GetTokenByID(app.Context.Config.AppConfig.ID, token)
		if err != nil && !errors.Is(err, appsecret.ErrTokenNotFound) {
			return nil, "", err
		}
		if tokenModel != nil {
			unmaskedSecrets = tokenModel.Secrets
		}
	}
	secretConfig, err := model.NewSecretConfig(cfg, unmaskedSecrets, now)
	if err != nil {
		return nil, "", err
	}

	return secretConfig, checksum, nil
}

// GenerateSecretVisitToken does not need connection.
func (s *AppService) GenerateSecretVisitToken(
	app *model.App,
	sessionInfo *apimodel.SessionInfo,
	visitingSecrets []config.SecretKey,
) (*appsecret.AppSecretVisitToken, error) {
	now := s.Clock.NowUTC()
	authenticatedAt := sessionInfo.AuthenticatedAt
	elapsed := now.Sub(authenticatedAt)
	if !(elapsed >= 0 && elapsed < 5*time.Minute || !sessionInfo.UserCanReauthenticate) {
		return nil, ErrReauthRequrired
	}

	token, err := s.AppSecretVisitTokenStore.CreateToken(
		app.Context.Config.AppConfig.ID,
		sessionInfo.UserID,
		visitingSecrets,
	)
	if err != nil {
		return nil, err
	}

	return token, nil
}

// GenerateTesterToken does not need connection.
func (s *AppService) GenerateTesterToken(
	app *model.App,
	returnURI string,
) (*tester.TesterToken, error) {
	return s.AppTesterTokenStore.CreateToken(config.AppID(app.ID), returnURI)
}

// Create calls other services that acquires connection themselves, and acquires connection.
func (s *AppService) Create(userID string, id string) (*model.App, error) {
	if err := s.validateAppID(id); err != nil {
		return nil, err
	}

	s.Logger.
		WithField("user_id", userID).
		WithField("app_id", id).
		Info("creating app")

	appHost, err := s.DefaultDomains.GetLatestAppHost(id)
	if err != nil {
		return nil, err
	}

	defaultAppPlan, err := s.Plan.GetDefaultPlan()
	if err != nil {
		return nil, err
	}

	createAppOpts, err := s.generateConfig(appHost, id, defaultAppPlan)
	if err != nil {
		return nil, err
	}

	err = s.GlobalDatabase.WithTx(func() error {
		err = s.AppConfigs.Create(createAppOpts)
		if err != nil {
			// TODO(portal): cleanup orphaned resources created from failed app creation
			s.Logger.WithError(err).WithField("app_id", id).Error("failed to create app")
			return err
		}

		err = s.DefaultDomains.CreateAllDefaultDomains(id)
		if err != nil {
			return err
		}

		err = s.AppAuthz.AddAuthorizedUser(id, userID, model.CollaboratorRoleOwner)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	app, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	return app, nil
}

// UpdateResources acquires connection.
func (s *AppService) UpdateResources(app *model.App, updates []appresource.Update) error {
	appResMgr := s.AppResMgrFactory.NewManagerWithAppContext(app.Context)
	files, err := appResMgr.ApplyUpdates(app.ID, updates)
	if err != nil {
		return err
	}

	err = s.GlobalDatabase.WithTx(func() error {
		return s.AppConfigs.UpdateResources(app.ID, files)
	})
	if err != nil {
		return err
	}

	return nil
}

// UpdateResources0 assumes acquired connection.
func (s *AppService) UpdateResources0(app *model.App, updates []appresource.Update) error {
	appResMgr := s.AppResMgrFactory.NewManagerWithAppContext(app.Context)
	files, err := appResMgr.ApplyUpdates(app.ID, updates)
	if err != nil {
		return err
	}

	err = s.AppConfigs.UpdateResources(app.ID, files)
	if err != nil {
		return err
	}

	return nil
}

func (s *AppService) generateResources(appHost string, appID string, featureConfig *config.FeatureConfig) (map[string][]byte, error) {
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
	createdAt := s.Clock.NowUTC()
	secretConfig := config.GenerateSecretConfigFromOptions(&config.GenerateSecretConfigOptions{}, createdAt, corerand.SecureRand)
	secretConfigYAML, err := yaml.Marshal(secretConfig)
	if err != nil {
		return nil, err
	}
	appResources[configsource.AuthgearSecretYAML] = secretConfigYAML

	// Assign feature config if any
	if featureConfig != nil {
		featureConfigYAML, err := yaml.Marshal(featureConfig)
		if err != nil {
			return nil, err
		}
		appResources[configsource.AuthgearFeatureYAML] = featureConfigYAML
	}

	// Generate translation json with default app name
	defaultTranslationJSONPath := path.Join(
		"templates", intl.BuiltinBaseLanguage, template.TranslationJSONName,
	)
	translationJSONObj := map[string]string{
		"app.name":              appID,
		"terms-of-service-link": DefaultTermsOfServiceLink,
		"privacy-policy-link":   DefaultPrivacyPolicyLink,
	}
	translationJSON, err := json.Marshal(translationJSONObj)
	if err != nil {
		return nil, err
	}
	appResources[defaultTranslationJSONPath] = translationJSON

	return appResources, nil
}

func (s *AppService) generateConfig(appHost string, appID string, appPlan *model.Plan) (opts *CreateAppOptions, err error) {
	appIDRegex, err := regexp.Compile(s.AppConfig.IDPattern)
	if err != nil {
		err = fmt.Errorf("invalid app ID validation pattern: %w", err)
		return
	}
	if !appIDRegex.MatchString(appID) {
		err = ErrAppIDInvalid
		return
	}

	var featureConfig *config.FeatureConfig
	planName := ""
	if appPlan != nil {
		featureConfig = appPlan.RawFeatureConfig
		planName = appPlan.Name
	}
	files, err := s.generateResources(appHost, appID, featureConfig)
	if err != nil {
		return
	}

	fs := afero.NewMemMapFs()
	for p, data := range files {
		_ = fs.MkdirAll(path.Dir(p), 0777)
		_ = afero.WriteFile(fs, p, data, 0666)
	}

	appFs := resource.LeveledAferoFs{Fs: fs, FsLevel: resource.FsLevelApp}
	appResMgr := s.AppResMgrFactory.NewManagerWithNewAppFS(appFs)
	_, err = appResMgr.ApplyUpdates(appID, nil)
	if err != nil {
		return
	}

	opts = &CreateAppOptions{
		AppID:     appID,
		Resources: files,
		PlanName:  planName,
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

func (s *AppService) RenderSAMLEntityID(appID string) string {
	return saml.RenderSAMLEntityID(s.SAMLEnvironmentConfig, appID)
}
