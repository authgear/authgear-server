package api

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/session"
)

type JSONResponseWriter interface {
	WriteResponse(rw http.ResponseWriter, resp *api.Response)
}

type RequireAuthenticatedMiddleware struct {
	JSON JSONResponseWriter
}

func (m RequireAuthenticatedMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := session.GetUserID(r.Context())
		if userID == nil {
			m.JSON.WriteResponse(w, &api.Response{Error: apierrors.NewUnauthorized("authentication required")})
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
