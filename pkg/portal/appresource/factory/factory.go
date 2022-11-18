package factory

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/hook"
	"github.com/authgear/authgear-server/pkg/lib/tutorial"
	"github.com/authgear/authgear-server/pkg/portal/appresource"
	"github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

type ManagerFactory struct {
	Context          context.Context
	AppBaseResources deps.AppBaseResources
	Tutorials        *tutorial.Service
	DenoClient       *hook.DenoClient
	Clock            clock.Clock
}

func (f *ManagerFactory) NewManagerWithAppContext(appContext *config.AppContext) *appresource.Manager {
	return &appresource.Manager{
		Context:            f.Context,
		AppResourceManager: appContext.Resources,
		AppFS:              appContext.AppFs,
		AppFeatureConfig:   appContext.Config.FeatureConfig,
		Tutorials:          f.Tutorials,
		DenoClient:         f.DenoClient,
		Clock:              f.Clock,
	}
}

func (f *ManagerFactory) NewManagerWithNewAppFS(appFs resource.Fs) *appresource.Manager {
	resMgr := (*resource.Manager)(f.AppBaseResources).Overlay(appFs)
	return &appresource.Manager{
		Context:            f.Context,
		AppResourceManager: resMgr,
		AppFS:              appFs,
		// The newly generated config should not violate any app plan
		// use default unlimited feature config for the app creation
		AppFeatureConfig: config.NewEffectiveDefaultFeatureConfig(),
		Tutorials:        f.Tutorials,
		DenoClient:       f.DenoClient,
		Clock:            f.Clock,
	}
}
