package middleware

import (
	"fmt"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/config"
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

		// Tenant configuration
		config.SetTenantConfig(r, &configuration)
		next.ServeHTTP(w, r)
	})
}
