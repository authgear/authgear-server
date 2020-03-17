package standalone

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/gateway/model"
	"github.com/skygeario/skygear-server/pkg/gateway/store"
)

type Store struct {
	TenantConfig config.TenantConfiguration
}

func (s *Store) GetDomain(domain string) (*model.Domain, error) {
	d := &model.Domain{}
	d.Assignment = model.AssignmentTypeMicroservices
	d.AppID = s.TenantConfig.AppID
	return d, nil
}

func (s *Store) GetDefaultDomain(domain string) (*model.Domain, error) {
	return nil, nil
}

func (s *Store) GetApp(id string) (*model.App, error) {
	app := &model.App{}
	app.ID = s.TenantConfig.AppID
	app.Name = s.TenantConfig.AppName
	app.Config = s.TenantConfig
	app.Plan = model.Plan{
		AuthEnabled: true,
	}
	app.AuthVersion = model.LiveVersion
	return app, nil
}

func (s *Store) GetLastDeploymentRoutes(app model.App) ([]*model.DeploymentRoute, error) {
	var routes []*model.DeploymentRoute
	for _, route := range s.TenantConfig.DeploymentRoutes {
		routes = append(routes, &model.DeploymentRoute{
			Version:    route.Version,
			Path:       route.Path,
			Type:       route.Type,
			TypeConfig: route.TypeConfig,
		})
	}
	return routes, nil
}

func (s *Store) GetLastDeploymentHooks(app model.App) (*model.DeploymentHooks, error) {
	var hooks = model.DeploymentHooks{
		AppID:            app.ID,
		IsLastDeployment: true,
	}
	for _, hook := range s.TenantConfig.Hooks {
		hooks.Hooks = append(hooks.Hooks, model.DeploymentHook{
			Event: hook.Event,
			URL:   hook.URL,
		})
	}
	return &hooks, nil
}

func (s *Store) Close() error {
	return nil
}

var (
	_ store.GatewayStore = &Store{}
)
