package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/web"
	"github.com/authgear/authgear-server/pkg/util/base32"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/rand"
)

// CSPNonceCookieDef is a HTTP session cookie.
// The nonce has to be stable within a browsing session because
// Turbo uses XHR to load new pages.
// If nonce changes on every page load, the script in the new page
// cannot be run in the current page due to different nonce.
var CSPNonceCookieDef = &httputil.CookieDef{
	NameSuffix: "csp_nonce",
	Path:       "/",
	SameSite:   http.SameSiteNoneMode,
}

type CSPNonceMiddleware struct {
	Cookies CookieManager
}

func (m *CSPNonceMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var nonce string
		cookie, err := m.Cookies.GetCookie(r, CSPNonceCookieDef)
		if err == nil {
			nonce = cookie.Value
		} else {
			nonce = rand.StringWithAlphabet(32, base32.Alphabet, rand.SecureRand)
			cookie := m.Cookies.ValueCookie(CSPNonceCookieDef, nonce)
			httputil.UpdateCookie(w, cookie)
		}
		r = r.WithContext(web.WithCSPNonce(r.Context(), nonce))
		next.ServeHTTP(w, r)
	})
}
