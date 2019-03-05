package middleware

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	gatewayConfig "github.com/skygeario/skygear-server/pkg/gateway/config"
	"github.com/skygeario/skygear-server/pkg/gateway/db"
	gatewayModel "github.com/skygeario/skygear-server/pkg/gateway/model"
)

// TenantAuthzMiddleware is middleware to check if the current app can access
// gear
type TenantAuthzMiddleware struct {
	Store        db.GatewayStore
	RouterConfig gatewayConfig.RouterConfig
}

// Handle reject the request if the current app doesn't have permission to
// access gear
func (a TenantAuthzMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app := gatewayModel.AppFromContext(r.Context())

		// Tenant authorization
		gear := gatewayModel.Gear(mux.Vars(r)["gear"])
		gearVersion := app.GetGearVersion(gear)
		if !app.CanAccessGear(gear) {
			http.Error(w, fmt.Sprintf("%s is not support in current app plan", gear), http.StatusForbidden)
			return
		}

		url, err := a.RouterConfig.GetGearURL(gear, gearVersion)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if url == "" {
			http.Error(w, fmt.Sprintf("%s gear %s environment is not supported", gear, gearVersion), http.StatusVariantAlsoNegotiates)
			return
		}

		r.Header.Set("X-Skygear-Gear", string(gear))
		r.Header.Set("X-Skygear-Gear-Version", string(gearVersion))
		r.Header.Set("X-Skygear-Gear-Endpoint", url)

		next.ServeHTTP(w, r)
	})
}
