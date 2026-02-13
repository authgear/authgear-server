package webapp

import (
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type RequireAuthenticatedMiddleware struct{}

func (m RequireAuthenticatedMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID := session.GetUserID(ctx)
		if userID == nil {
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
			httputil.Redirect(ctx, w, r, u.String(), http.StatusFound)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
