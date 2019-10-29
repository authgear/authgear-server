package provider

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/config"
	coreMiddleware "github.com/skygeario/skygear-server/pkg/core/middleware"
	"github.com/skygeario/skygear-server/pkg/gateway/model"
	"github.com/skygeario/skygear-server/pkg/gateway/store"
)

// GatewayTenantConfigurationProvider provide tenlent config from request
type GatewayTenantConfigurationProvider struct {
	coreMiddleware.ConfigurationProvider
	Store store.GatewayStore
}

// ProvideConfig function query the tenant config from db by request
func (p GatewayTenantConfigurationProvider) ProvideConfig(r *http.Request) (config.TenantConfiguration, error) {
	ctx := model.GatewayContextFromContext(r.Context())
	app := ctx.App
	if app.ID == "" {
		panic("config provider: app not found")
	}

	return app.Config, nil
}
