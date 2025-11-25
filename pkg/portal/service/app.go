package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
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
	"github.com/authgear/authgear-server/pkg/lib/config/plan"
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
	corerand "github.com/authgear/authgear-server/pkg/util/rand"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
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

var projectIDPattern_valid = regexp.MustCompile("^[a-z0-9][a-z0-9-]{2,30}[a-z0-9]$")
var projectIDPattern_invalid = regexp.MustCompile("^[a-z]{2}-")

type AppConfigService interface {
	ResolveContext(ctx context.Context, appID string, fn func(context.Context, *config.AppContext) error) error
	UpdateResources(ctx context.Context, appID string, updates []*resource.ResourceFile) error
	Create(ctx context.Context, opts *CreateAppOptions) error
}

type AppAuthzService interface {
	AddAuthorizedUser(ctx context.Context, appID string, userID string, role model.CollaboratorRole) error
	ListAuthorizedApps(ctx context.Context, userID string) ([]string, error)
}

type AppDefaultDomainService interface {
	GetLatestAppHost(appID string) (string, error)
	CreateAllDefaultDomains(ctx context.Context, appID string) error
}

type AppPlanService interface {
	GetDefaultPlan(ctx context.Context) (*plan.Plan, error)
}

var AppServiceLogger = slogutil.NewLogger("app-service")

type AppResourceManagerFactory interface {
	NewManagerWithNewAppFS(appFs resource.Fs) *appresource.Manager
	NewManagerWithAppContext(appContext *config.AppContext) *appresource.Manager
}

type AppSecretVisitTokenStore interface {
	CreateToken(
		ctx context.Context,
		appID config.AppID,
		userID string,
		secrets []config.SecretKey,
	) (*appsecret.AppSecretVisitToken, error)
	GetTokenByID(
		ctx context.Context,
		appID config.AppID,
		tokenID string,
	) (*appsecret.AppSecretVisitToken, error)
}

type AppTesterTokenStore interface {
	CreateToken(
		ctx context.Context,
		appID config.AppID,
		returnURI string,
	) (*tester.TesterToken, error)
}

type AppConfigSourceStore interface {
	ListAll(ctx context.Context) ([]*configsource.DatabaseSource, error)
}

type AppService struct {
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
	ConfigSourceStore        AppConfigSourceStore
	DatabaseCredentials      *config.GlobalDatabaseCredentialsEnvironmentConfig
	RedisCredentials         *config.GlobalRedisCredentialsEnvironmentConfig
}

// Get calls other services that acquires connection themselves.
func (s *AppService) Get(ctx context.Context, id string) (*model.App, error) {
	var appCtx *config.AppContext
	err := s.AppConfigs.ResolveContext(ctx, id, func(ctx context.Context, ac *config.AppContext) error {
		appCtx = ac
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &model.App{
		ID:      id,
		Context: appCtx,
	}, nil
}

// GetMany just uses Get.
func (s *AppService) GetMany(ctx context.Context, ids []string) (out []*model.App, err error) {
	for _, id := range ids {
		app, err := s.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		out = append(out, app)
	}

	return
}

// GetAppList calls other services that acquires connection themselves.
func (s *AppService) GetAppList(ctx context.Context, userID string) ([]*model.AppListItem, error) {
	appIDs, err := s.AppAuthz.ListAuthorizedApps(ctx, userID)
	if err != nil {
		return nil, err
	}

	apps, err := s.GetMany(ctx, appIDs)
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
func (s *AppService) GetProjectQuota(ctx context.Context, userID string) (int, error) {
	q := s.SQLBuilder.Select("max_own_apps").
		From(s.SQLBuilder.TableName("_portal_user_app_quota")).
		Where("user_id = ?", userID)

	var quota int
	err := s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		row, err := s.SQLExecutor.QueryRowWith(ctx, q)
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
func (s *AppService) GetManyProjectQuota(ctx context.Context, userIDs []string) ([]int, error) {
	q := s.SQLBuilder.Select(
		"user_id",
		"max_own_apps",
	).
		From(s.SQLBuilder.TableName("_portal_user_app_quota")).
		Where("user_id = ANY (?)", pq.Array(userIDs))

	m := make(map[string]int)
	err := s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		rows, err := s.SQLExecutor.QueryWith(ctx, q)
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
func (s *AppService) LoadRawAppConfig(ctx context.Context, app *model.App) (*config.AppConfig, string, error) {
	resMgr := s.AppResMgrFactory.NewManagerWithAppContext(app.Context)
	result, err := resMgr.ReadAppFile(ctx, configsource.AppConfig,
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
	ctx context.Context,
	app *model.App,
	sessionInfo *apimodel.SessionInfo,
	token string) (*model.SecretConfig, string, error) {
	var unmaskedSecrets []config.SecretKey = []config.SecretKey{}
	resMgr := s.AppResMgrFactory.NewManagerWithAppContext(app.Context)
	result, err := resMgr.ReadAppFile(ctx, configsource.SecretConfig, &resource.AppFile{
		Path: configsource.AuthgearSecretYAML,
	})
	if err != nil {
		return nil, "", err
	}

	bytes := result.([]byte)
	checksum := checksum.CRC32IEEEInHex(bytes)

	cfg, err := config.ParsePartialSecret(ctx, bytes)
	if err != nil {
		return nil, "", err
	}

	now := s.Clock.NowUTC()
	if token != "" {
		tokenModel, err := s.AppSecretVisitTokenStore.GetTokenByID(ctx, app.Context.Config.AppConfig.ID, token)
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

// LoadEffectiveSecretConfig does not need connection.
func (s *AppService) LoadEffectiveSecretConfig(ctx context.Context, app *model.App) (*model.EffectiveSecretConfig, error) {
	resMgr := s.AppResMgrFactory.NewManagerWithAppContext(app.Context)
	result, err := resMgr.ReadEffectiveFile(ctx, configsource.SecretConfig, &resource.EffectiveResource{})
	if err != nil {
		return nil, err
	}
	secretConfig, ok := result.(*config.SecretConfig)
	if !ok {
		panic(fmt.Errorf("unexpected: result is not a SecretConfig"))
	}

	effectiveSecretConfig := &model.EffectiveSecretConfig{
		OAuthSSOProviderDemoSecrets: []model.OAuthSSOProviderDemoSecretItem{},
	}

	demoSecrets, ok := secretConfig.LookupData(config.SSOOAuthDemoCredentialsKey).(*config.SSOOAuthDemoCredentials)
	if !ok {
		return effectiveSecretConfig, nil
	}

	for _, item := range demoSecrets.Items {
		item := item
		effectiveSecretConfig.OAuthSSOProviderDemoSecrets = append(effectiveSecretConfig.OAuthSSOProviderDemoSecrets, model.OAuthSSOProviderDemoSecretItem{
			Type: item.ProviderConfig.Type(),
		})
	}

	return effectiveSecretConfig, nil
}

// GenerateSecretVisitToken does not need connection.
func (s *AppService) GenerateSecretVisitToken(
	ctx context.Context,
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
		ctx,
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
	ctx context.Context,
	app *model.App,
	returnURI string,
) (*tester.TesterToken, error) {
	return s.AppTesterTokenStore.CreateToken(ctx, config.AppID(app.ID), returnURI)
}

// Create calls other services that acquires connection themselves, and acquires connection.
func (s *AppService) Create(ctx context.Context, userID string, id string) (*model.App, error) {
	logger := AppServiceLogger.GetLogger(ctx)
	if err := s.validateAppID(ctx, id); err != nil {
		return nil, err
	}

	logger.Info(ctx, "creating app", slog.String("user_id", userID), slog.String("app_id", id))

	appHost, err := s.DefaultDomains.GetLatestAppHost(id)
	if err != nil {
		return nil, err
	}

	defaultAppPlan, err := s.Plan.GetDefaultPlan(ctx)
	if err != nil {
		return nil, err
	}

	err = s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		createAppOpts, err := s.generateConfig(ctx, appHost, id, defaultAppPlan)
		if err != nil {
			return err
		}
		err = s.AppConfigs.Create(ctx, createAppOpts)
		if err != nil {
			if errors.Is(err, ErrDuplicatedAppID) {
				logger.Info(ctx, "failed to create duplicated app", slog.String("app_id", id))
				return err
			} else {
				// TODO(portal): cleanup orphaned resources created from failed app creation
				logger.WithError(err).Error(ctx, "failed to create app", slog.String("app_id", id))
				return err
			}
		}

		err = s.DefaultDomains.CreateAllDefaultDomains(ctx, id)
		if err != nil {
			return err
		}

		err = s.AppAuthz.AddAuthorizedUser(ctx, id, userID, model.CollaboratorRoleOwner)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	app, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return app, nil
}

// UpdateResources acquires connection.
func (s *AppService) UpdateResources(ctx context.Context, app *model.App, updates []appresource.Update) error {
	type arity1ReturningError func(ctx context.Context) error

	// applyUpdatesToTheOriginalApp DOES NOT reference to the original arguments.
	// So it can be used to apply updates to other apps.
	applyUpdatesToGivenApp := func(app *model.App, updates []appresource.Update) arity1ReturningError {
		return func(ctx context.Context) error {
			appResMgr := s.AppResMgrFactory.NewManagerWithAppContext(app.Context)
			files, err := appResMgr.ApplyUpdates0(ctx, app.ID, updates)
			if err != nil {
				return err
			}
			return s.AppConfigs.UpdateResources(ctx, app.ID, files)
		}
	}

	funcs := []arity1ReturningError{
		applyUpdatesToGivenApp(app, updates),
	}

	if AUTHGEARONCE {
		smtpUpdate, err := s.makeSMTPUpdate(updates)
		if err != nil {
			return err
		}
		if smtpUpdate != nil {
			logger := AppServiceLogger.GetLogger(ctx)
			logger.Info(ctx, "detected SMTP secret update", slog.String("source_app_id", app.ID))
			appIDsToPropagate, err := s.getAllAppIDsExcept(ctx, app.ID)
			if err != nil {
				return err
			}

			theUpdates := []appresource.Update{*smtpUpdate}
			for _, appID := range appIDsToPropagate {
				theApp, err := s.Get(ctx, appID)
				if err != nil {
					return err
				}

				logger.Info(ctx, "propagate STMP secret update", slog.String("source_app_id", app.ID), slog.String("target_app_id", theApp.ID))
				funcs = append(funcs, applyUpdatesToGivenApp(theApp, theUpdates))
			}
		}
	}

	var err error
	err = s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		for _, f := range funcs {
			err := f(ctx)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// UpdateResources0 assumes acquired connection.
func (s *AppService) UpdateResources0(ctx context.Context, app *model.App, updates []appresource.Update) error {
	appResMgr := s.AppResMgrFactory.NewManagerWithAppContext(app.Context)
	files, err := appResMgr.ApplyUpdates0(ctx, app.ID, updates)
	if err != nil {
		return err
	}

	err = s.AppConfigs.UpdateResources(ctx, app.ID, files)
	if err != nil {
		return err
	}

	return nil
}

func (s *AppService) makeSMTPUpdate(updates []appresource.Update) (out *appresource.Update, err error) {
	for _, update := range updates {
		if update.Path == configsource.AuthgearSecretYAML {
			var instructions *config.SecretConfigUpdateInstruction
			instructions, err = configsource.ParseAuthgearSecretsYAMLUpdateInstructions(update.Data)
			if err != nil {
				return

			}

			if instructions.SMTPServerCredentialsUpdateInstruction != nil {
				syntheticInstructions := &config.SecretConfigUpdateInstruction{
					SMTPServerCredentialsUpdateInstruction: instructions.SMTPServerCredentialsUpdateInstruction,
				}

				var data []byte
				data, err = json.Marshal(syntheticInstructions)
				if err != nil {
					return
				}

				return &appresource.Update{
					Path: update.Path,
					Data: data,
				}, nil
			}
		}
	}
	return
}

// getAllAppIDsExcept acquires connection.
func (s *AppService) getAllAppIDsExcept(ctx context.Context, exceptAppID string) ([]string, error) {
	var allAppIDs []string
	err := s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		srcs, err := s.ConfigSourceStore.ListAll(ctx)
		if err != nil {
			return err
		}
		for _, src := range srcs {
			if src.AppID != exceptAppID {
				allAppIDs = append(allAppIDs, src.AppID)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return allAppIDs, nil
}

func (s *AppService) generateResources(appHost string, appID string) (map[string][]byte, error) {
	appResources := make(map[string][]byte)

	// Generate app config
	publicOrigin := &url.URL{Scheme: "https", Host: appHost}
	appConfig := config.GenerateAppConfigFromOptions(&config.GenerateAppConfigOptions{
		AppID:                       appID,
		PublicOrigin:                publicOrigin.String(),
		CookieDomain:                appHost,
		PhoneNumberValidationMethod: config.LibphonenumberValidationMethodIsPossibleNumber,
	})
	appConfigYAML, err := yaml.Marshal(appConfig)
	if err != nil {
		return nil, err
	}
	appResources[configsource.AuthgearYAML] = appConfigYAML

	// Generate secret config
	createdAt := s.Clock.NowUTC()
	secretOpts := &config.GenerateSecretConfigOptions{}

	// Include database and Redis credentials from portal environment
	// These are shared across all projects in a Helm deployment
	if s.DatabaseCredentials != nil {
		secretOpts.DatabaseURL = s.DatabaseCredentials.DatabaseURL
		secretOpts.DatabaseSchema = s.DatabaseCredentials.DatabaseSchema
	}
	if s.RedisCredentials != nil {
		secretOpts.RedisURL = s.RedisCredentials.RedisURL
	}

	secretConfig := config.GenerateSecretConfigFromOptions(secretOpts, createdAt, corerand.SecureRand)
	secretConfigYAML, err := yaml.Marshal(secretConfig)
	if err != nil {
		return nil, err
	}
	appResources[configsource.AuthgearSecretYAML] = secretConfigYAML

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

func (s *AppService) generateConfig(ctx context.Context, appHost string, appID string, appPlan *plan.Plan) (opts *CreateAppOptions, err error) {
	planName := ""
	if appPlan != nil {
		planName = appPlan.Name
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
	appResMgr := s.AppResMgrFactory.NewManagerWithNewAppFS(appFs)
	_, err = appResMgr.ApplyUpdates0(ctx, appID, nil)
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

func (s *AppService) validateAppID(ctx context.Context, appID string) error {
	if !projectIDPattern_valid.MatchString(appID) {
		return ErrAppIDInvalid
	}

	if projectIDPattern_invalid.MatchString(appID) {
		// It is hard to explain this pattern to the developers.
		// So let's just say the reason is reserved.
		return ErrAppIDReserved
	}

	var list *blocklist.Blocklist

	result, err := s.Resources.Read(ctx, portalresource.ReservedAppIDTXT, resource.EffectiveResource{})
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

func (s *AppService) LoadAppWebhookSecretMaterials(
	ctx context.Context,
	app *model.App) (*config.WebhookKeyMaterials, error) {
	resMgr := s.AppResMgrFactory.NewManagerWithAppContext(app.Context)
	result, err := resMgr.ReadAppFile(ctx, configsource.SecretConfig, &resource.AppFile{
		Path: configsource.AuthgearSecretYAML,
	})
	if err != nil {
		return nil, nil
	}

	bytes := result.([]byte)

	secretConfig, err := config.ParsePartialSecret(ctx, bytes)
	if err != nil {
		return nil, err
	}
	if webhook, ok := secretConfig.LookupData(config.WebhookKeyMaterialsKey).(*config.WebhookKeyMaterials); ok {
		return webhook, nil
	}

	return nil, fmt.Errorf("unexpected: app %s is missing webhook secret", app.ID)
}
