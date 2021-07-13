package factory

import (
	"github.com/authgear/authgear-server/pkg/portal/appresource"
	"github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

type ManagerFactory struct {
	AppBaseResources deps.AppBaseResources
}

func (f *ManagerFactory) NewManagerWithApp(app *model.App) *appresource.Manager {
	return &appresource.Manager{
		AppResourceManager: app.Context.Resources,
		AppFS:              app.Context.AppFs,
	}
}

func (f *ManagerFactory) NewManagerWithNewAppFS(appFs resource.Fs) *appresource.Manager {
	resMgr := (*resource.Manager)(f.AppBaseResources).Overlay(appFs)
	return &appresource.Manager{
		AppResourceManager: resMgr,
		AppFS:              appFs,
	}
}
