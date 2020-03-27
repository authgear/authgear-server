package middleware

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/authn"
)

// AuthnMiddleware populate auth context information by reading request headers
type AuthnMiddleware struct {
}

func (m AuthnMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authninfo, err := authn.ParseHeaders(r)
		if err != nil {
			panic(err)
		}

		if authninfo != nil {
			if authninfo.IsValid {
				r = r.WithContext(authn.WithAuthn(r.Context(), authninfo, authninfo.User()))
			} else {
				r = r.WithContext(authn.WithInvalidAuthn(r.Context()))
			}
		}

		next.ServeHTTP(w, r)
	})
}
