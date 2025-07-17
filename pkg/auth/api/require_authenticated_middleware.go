package api

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type RequireAuthenticatedMiddleware struct{}

func (m RequireAuthenticatedMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID := session.GetUserID(ctx)
		if userID == nil {
			httputil.WriteJSONResponse(ctx, w, &api.Response{Error: apierrors.NewUnauthorized("authentication required")})
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
