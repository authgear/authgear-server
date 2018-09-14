package middleware

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/skygeario/skygear-server/pkg/gateway/db"
	gatewayModel "github.com/skygeario/skygear-server/pkg/gateway/model"
)

// TenantAuthzMiddleware is middleware to check if the current app can access
// gear
type TenantAuthzMiddleware struct {
	Store db.GatewayStore
}

// Handle reject the request if the current app doesn't have permission to
// access gear
func (a TenantAuthzMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		app := gatewayModel.App{}
		if err := a.Store.GetAppByDomain(host, &app); err != nil {
			http.Error(w, "Fail to found app", http.StatusBadRequest)
			return
		}

		// Tenant authorization
		gear := mux.Vars(r)["gear"]
		if !app.CanAccessGear(gear) {
			http.Error(w, fmt.Sprintf("%s is not support in current app plan", gear), http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
