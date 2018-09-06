package middleware

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/model"
)

type EnvTenantMiddleware struct{}

func (m EnvTenantMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		configuration := config.TenantConfiguration{}
		configuration.ReadFromEnv()

		// Tenant authentication
		// Set key type to header only, no rejection
		apiKey := model.GetAPIKey(r)
		apiKeyType := model.CheckAccessKeyType(configuration, apiKey)
		model.SetAccessKeyType(r, apiKeyType)

		// Tenant configuration
		config.SetTenantConfig(r, configuration)
		next.ServeHTTP(w, r)
	})
}
