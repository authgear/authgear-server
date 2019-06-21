package middleware

import (
	"net/http"

	gatewayModel "github.com/skygeario/skygear-server/pkg/gateway/model"
	"github.com/skygeario/skygear-server/pkg/gateway/store"
)

type FindAppMiddleware struct {
	Store store.GatewayStore
}

func (f FindAppMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		app := gatewayModel.App{}
		if err := f.Store.GetAppByDomain(host, &app); err != nil {
			http.Error(w, "Fail to found app", http.StatusBadRequest)
			return
		}

		ctx := gatewayModel.GatewayContextFromContext(r.Context())
		ctx.App = app
		r = r.WithContext(gatewayModel.ContextWithGatewayContext(r.Context(), ctx))

		next.ServeHTTP(w, r)
	})
}
