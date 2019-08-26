package middleware

import (
	"fmt"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/model"
)

type ConfigurationProvider interface {
	ProvideConfig(r *http.Request) (config.TenantConfiguration, error)
}

type ConfigurationProviderFunc func(r *http.Request) (config.TenantConfiguration, error)

func (f ConfigurationProviderFunc) ProvideConfig(r *http.Request) (config.TenantConfiguration, error) {
	return f(r)
}

type TenantConfigurationMiddleware struct {
	ConfigurationProvider
}

func (m TenantConfigurationMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		configuration, err := m.ProvideConfig(r)
		if err != nil {
			panic(fmt.Errorf("Unable to retrieve configuration: %v", err.Error()))
		}

		// FIXME(middleware):
		// This will be overwritten by core.server, should refactor this.
		// For now, set checked access key to header and let core.server read it from header.
		r = auth.InitRequestAuthContext(r)
		authCtx := auth.NewContextSetterWithContext(r.Context())

		// Tenant authentication
		// Set access key to header only, no rejection
		apiKey := model.GetAPIKey(r)
		accessKey := model.CheckAccessKey(configuration, apiKey)
		model.SetAccessKey(r, accessKey)
		authCtx.SetAccessKey(accessKey)

		// Tenant configuration
		config.SetTenantConfig(r, &configuration)
		next.ServeHTTP(w, r)
	})
}
