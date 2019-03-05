package provider

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/config"
	coreMiddleware "github.com/skygeario/skygear-server/pkg/core/middleware"
	"github.com/skygeario/skygear-server/pkg/gateway/db"
	"github.com/skygeario/skygear-server/pkg/gateway/model"
)

// GatewayTenantConfigurationProvider provide tenlent config from request
type GatewayTenantConfigurationProvider struct {
	coreMiddleware.ConfigurationProvider
	Store db.GatewayStore
}

// ProvideConfig function query the tenant config from db by request
func (p GatewayTenantConfigurationProvider) ProvideConfig(r *http.Request) (config.TenantConfiguration, error) {
	app := model.AppFromContext(r.Context())
	if app == nil {
		panic("Unexpected app not found")
	}

	return app.Config, nil
}
