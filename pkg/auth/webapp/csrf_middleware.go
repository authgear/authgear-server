package webapp

import (
	"net/http"

	"github.com/gorilla/csrf"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/jwkutil"
)

type CSRFMiddleware struct {
	Secret     *config.CSRFKeyMaterials
	CookieDef  CSRFCookieDef
	TrustProxy config.TrustProxy
}

func (m *CSRFMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secure := httputil.GetProto(r, bool(m.TrustProxy)) == "https"
		options := []csrf.Option{
			csrf.CookieName(m.CookieDef.Name),
			csrf.Path("/"),
			csrf.Secure(secure),
		}
		if m.CookieDef.Domain != "" {
			options = append(options, csrf.Domain(m.CookieDef.Domain))
		}

		useragent := r.UserAgent()
		if httputil.ShouldSendSameSiteNone(useragent, secure) {
			options = append(options, csrf.SameSite(csrf.SameSiteNoneMode))
		} else {
			// http.Cookie SameSiteDefaultMode option will write SameSite
			// with empty value to the cookie header which doesn't work for
			// some old browsers
			// ref: https://github.com/golang/go/issues/36990
			// To avoid writing samesite to the header
			// set empty value to Cookie SameSite
			// https://golang.org/src/net/http/cookie.go#L220
			options = append(options, csrf.SameSite(0))
		}

		key, err := jwkutil.ExtractOctetKey(&m.Secret.Set, "")
		if err != nil {
			panic("webapp: CSRF key not found")
		}
		gorillaCSRF := csrf.Protect(key, options...)
		h := gorillaCSRF(next)
		h.ServeHTTP(w, r)
	})
}
