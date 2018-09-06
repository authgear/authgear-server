package middleware

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/model"
	gatewayModel "github.com/skygeario/skygear-server/pkg/gateway/model"
)

type TenantMiddleware struct {
}

func (a TenantMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		app := gatewayModel.GetApp(host)
		if app == nil {
			http.Error(w, "App not found", http.StatusNotFound)
			return
		}

		// Tenant authentication
		// Set key type to header only, no rejection
		apiKey := model.GetAPIKey(r)
		apiKeyType := model.CheckAccessKeyType(app.Config, apiKey)
		model.SetAccessKeyType(r, apiKeyType)

		// Tenant authorization
		gear := mux.Vars(r)["gear"]
		if !app.CanAccessGear(gear) {
			http.Error(w, fmt.Sprintf("%s is not support in current app plan", gear), http.StatusForbidden)
			return
		}

		// Tenant configuration
		config.SetTenantConfig(r, app.Config)
		next.ServeHTTP(w, r)
	})
}
