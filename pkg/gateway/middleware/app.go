package middleware

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/gateway/db"
	gatewayModel "github.com/skygeario/skygear-server/pkg/gateway/model"
)

type FindAppMiddleware struct {
	Store db.GatewayStore
}

func (f FindAppMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		app := gatewayModel.App{}
		if err := f.Store.GetAppByDomain(host, &app); err != nil {
			http.Error(w, "Fail to found app", http.StatusBadRequest)
			return
		}

		r = r.WithContext(gatewayModel.ContextWithApp(r.Context(), &app))

		next.ServeHTTP(w, r)
	})
}
