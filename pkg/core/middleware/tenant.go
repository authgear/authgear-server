package middleware

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/errors"
)

type ConfigurationProvider interface {
	ProvideConfig(r *http.Request) (config.TenantConfiguration, error)
}

type ConfigurationProviderFunc func(r *http.Request) (config.TenantConfiguration, error)

func (f ConfigurationProviderFunc) ProvideConfig(r *http.Request) (config.TenantConfiguration, error) {
	return f(r)
}

type WriteTenantConfigMiddleware struct {
	ConfigurationProvider
}

func (m WriteTenantConfigMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		configuration, err := m.ProvideConfig(r)
		if err != nil {
			panic(errors.Newf("unable to retrieve configuration: %w", err))
		}
		config.WriteTenantConfig(r, &configuration)

		r = r.WithContext(config.WithTenantConfig(r.Context(), &configuration))
		next.ServeHTTP(w, r)
	})
}

type ReadTenantConfigMiddleware struct{}

func (m ReadTenantConfigMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		configuration := config.ReadTenantConfig(r)

		r = r.WithContext(config.WithTenantConfig(r.Context(), &configuration))
		next.ServeHTTP(w, r)
	})
}
