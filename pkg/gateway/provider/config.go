package provider

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/gateway/model"
)

func NewTenantConfigurationFromRequest(r *http.Request) (config.TenantConfiguration, error) {
	host := r.Host
	// TODO:
	// should return error if failed instead of panic?
	app := model.GetApp(host)
	return app.Config, nil
}
