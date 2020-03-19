package webapp

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
)

type RequiredAnonymousMiddleware struct{}

func (m RequiredAnonymousMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authInfo := auth.GetAuthInfo(r.Context())
		if authInfo != nil {
			// TODO(webapp): Respect redirect_uri
			http.Redirect(w, r, "/settings", http.StatusFound)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
