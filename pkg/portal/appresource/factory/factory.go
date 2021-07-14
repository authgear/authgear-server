package factory

import (
	"github.com/authgear/authgear-server/pkg/portal/appresource"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

type ManagerFactory struct {
	AppBaseResources   deps.AppBaseResources
	SecretKeyAllowlist portalconfig.SecretKeyAllowlist
}

func (f *ManagerFactory) NewManagerWithApp(app *model.App) *appresource.Manager {
	return &appresource.Manager{
		AppResourceManager: app.Context.Resources,
		AppFS:              app.Context.AppFs,
		SecretKeyAllowlist: f.SecretKeyAllowlist,
	}
}

func (f *ManagerFactory) NewManagerWithNewAppFS(appFs resource.Fs) *appresource.Manager {
	resMgr := (*resource.Manager)(f.AppBaseResources).Overlay(appFs)
	return &appresource.Manager{
		AppResourceManager: resMgr,
		AppFS:              appFs,
		SecretKeyAllowlist: f.SecretKeyAllowlist,
	}
}
