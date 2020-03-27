package middleware

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth"
)

// AccessKeyMiddleware populate access key context information by reading request headers
type AccessKeyMiddleware struct {
}

func (m AccessKeyMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		accessKey := m.resolve(r)
		r = r.WithContext(auth.WithAccessKey(r.Context(), accessKey))
		next.ServeHTTP(rw, r)
	})
}

func (m AccessKeyMiddleware) resolve(r *http.Request) auth.AccessKey {
	return *auth.NewAccessKeyFromRequest(r)
}
