package provider

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/logging"
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
	logger := logging.LoggerEntry("gateway")

	host := r.Host
	app := model.App{}
	err := p.Store.GetAppByDomain(host, &app)
	if err != nil {
		logger.WithError(err).Warn("Fail to found app")
	}
	return app.Config, err
}
