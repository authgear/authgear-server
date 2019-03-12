package middleware

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/skygeario/skygear-server/pkg/gateway/db"
	gatewayModel "github.com/skygeario/skygear-server/pkg/gateway/model"
)

type FindCloudCodeMiddleware struct {
	RestPathIdentifier string
	Store              db.GatewayStore
}

func (f FindCloudCodeMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := gatewayModel.GatewayContextFromContext(r.Context())
		app := ctx.App
		cloudCode := gatewayModel.CloudCode{}

		path := "/" + mux.Vars(r)[f.RestPathIdentifier]
		if err := f.Store.FindLongestMatchedCloudCode(path, app, &cloudCode); err != nil {
			http.Error(w, "Fail to found cloud code", http.StatusBadRequest)
			return
		}

		ctx.CloudCode = cloudCode
		r = r.WithContext(gatewayModel.ContextWithGatewayContext(r.Context(), ctx))

		next.ServeHTTP(w, r)
	})
}
