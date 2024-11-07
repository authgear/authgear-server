package factory

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/hook"
	"github.com/authgear/authgear-server/pkg/lib/tutorial"
	"github.com/authgear/authgear-server/pkg/portal/appresource"
	"github.com/authgear/authgear-server/pkg/portal/deps"
	portalservice "github.com/authgear/authgear-server/pkg/portal/service"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

type ManagerFactoryLogger struct{ *log.Logger }

func NewManagerFactoryLogger(lf *log.Factory) ManagerFactoryLogger {
	return ManagerFactoryLogger{lf.New("appresource-manager")}
}

type ManagerFactory struct {
	Logger            ManagerFactoryLogger
	AppBaseResources  deps.AppBaseResources
	Tutorials         *tutorial.Service
	DenoClient        *hook.DenoClientImpl
	Clock             clock.Clock
	EnvironmentConfig *config.EnvironmentConfig
	DomainService     *portalservice.DomainService
}

func (f *ManagerFactory) NewManagerWithAppContext(appContext *config.AppContext) *appresource.Manager {
	return &appresource.Manager{
		Logger:                f.Logger.Logger,
		AppResourceManager:    appContext.Resources,
		AppFS:                 appContext.AppFs,
		AppFeatureConfig:      appContext.Config.FeatureConfig,
		AppHostSuffixes:       &f.EnvironmentConfig.AppHostSuffixes,
		DomainService:         f.DomainService,
		Tutorials:             f.Tutorials,
		DenoClient:            f.DenoClient,
		Clock:                 f.Clock,
		SAMLEnvironmentConfig: f.EnvironmentConfig.SAML,
	}
}

func (f *ManagerFactory) NewManagerWithNewAppFS(appFs resource.Fs) *appresource.Manager {
	resMgr := (*resource.Manager)(f.AppBaseResources).Overlay(appFs)
	return &appresource.Manager{
		Logger:             f.Logger.Logger,
		AppResourceManager: resMgr,
		AppFS:              appFs,
		// The newly generated config should not violate any app plan
		// use default unlimited feature config for the app creation
		AppFeatureConfig:      config.NewEffectiveDefaultFeatureConfig(),
		AppHostSuffixes:       &f.EnvironmentConfig.AppHostSuffixes,
		DomainService:         f.DomainService,
		Tutorials:             f.Tutorials,
		DenoClient:            f.DenoClient,
		Clock:                 f.Clock,
		SAMLEnvironmentConfig: f.EnvironmentConfig.SAML,
	}
}
