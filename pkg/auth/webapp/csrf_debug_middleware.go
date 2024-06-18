package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func (m *CSRFDebugMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		omitCookie := m.Cookies.ValueCookie(CSRFDebugCookieSameSiteOmitDef, "exists")
		httputil.UpdateCookie(w, omitCookie)
		noneCookie := m.Cookies.ValueCookie(CSRFDebugCookieSameSiteNoneDef, "exists")
		httputil.UpdateCookie(w, noneCookie)
		laxCookie := m.Cookies.ValueCookie(CSRFDebugCookieSameSiteLaxDef, "exists")
		httputil.UpdateCookie(w, laxCookie)
		strictCookie := m.Cookies.ValueCookie(CSRFDebugCookieSameSiteStrictDef, "exists")
		httputil.UpdateCookie(w, strictCookie)
		next.ServeHTTP(w, r)
	})
}
