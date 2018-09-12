package middleware

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	gatewayModel "github.com/skygeario/skygear-server/pkg/gateway/model"
)

type TenantAuthzMiddleware struct {
}

func (a TenantAuthzMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		app := gatewayModel.GetApp(host)

		// Tenant authorization
		gear := mux.Vars(r)["gear"]
		if !app.CanAccessGear(gear) {
			http.Error(w, fmt.Sprintf("%s is not support in current app plan", gear), http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
