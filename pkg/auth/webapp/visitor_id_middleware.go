package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type VisitorIDMiddleware struct {
	Cookies CookieManager
}

func (m *VisitorIDMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var visitorID string
		cookie, err := m.Cookies.GetCookie(r, VisitorIDCookieDef)
		if err == nil {
			visitorID = cookie.Value
		} else {
			// create new visitor id
			visitorID = uuid.New()
			cookie := m.Cookies.ValueCookie(VisitorIDCookieDef, visitorID)
			httputil.UpdateCookie(w, cookie)
		}
		ctx := WithVisitorID(r.Context(), visitorID)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
