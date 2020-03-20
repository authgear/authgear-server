package webapp

import (
	"net/http"
	"net/url"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
)

type RequiredAuthenticatedMiddleware struct{}

func (m RequiredAuthenticatedMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authInfo := auth.GetAuthInfo(r.Context())
		if authInfo == nil {
			// Trim scheme and host, retain path and query
			redirectURI := *r.URL
			redirectURI.Scheme = ""
			redirectURI.Host = ""
			q := url.Values{}
			q.Set("redirect_uri", redirectURI.String())
			u := url.URL{
				Path:     "/",
				RawQuery: q.Encode(),
			}
			http.Redirect(w, r, u.String(), http.StatusFound)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
