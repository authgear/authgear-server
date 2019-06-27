package middleware

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	coreHttp "github.com/skygeario/skygear-server/pkg/core/http"
	gatewayConfig "github.com/skygeario/skygear-server/pkg/gateway/config"
	gatewayModel "github.com/skygeario/skygear-server/pkg/gateway/model"
	"github.com/skygeario/skygear-server/pkg/gateway/store"
)

// TenantAuthzMiddleware is middleware to check if the current app can access
// gear
type TenantAuthzMiddleware struct {
	Store         store.GatewayStore
	Configuration gatewayConfig.Configuration
}

// Handle reject the request if the current app doesn't have permission to
// access gear
func (a TenantAuthzMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := gatewayModel.GatewayContextFromContext(r.Context())
		app := ctx.App

		// Tenant authorization
		gear := gatewayModel.Gear(mux.Vars(r)["gear"])
		gearVersion := app.GetGearVersion(gear)
		if !app.CanAccessGear(gear) {
			http.Error(w, fmt.Sprintf("%s is not support in current app plan", gear), http.StatusForbidden)
			return
		}

		url, err := a.Configuration.GetGearURL(gear, gearVersion)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if url == "" {
			http.Error(w, fmt.Sprintf("%s gear %s environment is not supported", gear, gearVersion), http.StatusVariantAlsoNegotiates)
			return
		}

		r.Header.Set(coreHttp.HeaderGear, string(gear))
		r.Header.Set(coreHttp.HeaderGearVersion, string(gearVersion))
		r.Header.Set(coreHttp.HeaderGearEndpoint, url)

		next.ServeHTTP(w, r)
	})
}
