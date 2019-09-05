package middleware

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth"
)

// AuthMiddleware setup auth context in request
type AuthMiddleware struct {
}

func (m AuthMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = auth.InitRequestAuthContext(r)
		next.ServeHTTP(w, r)
	})
}
