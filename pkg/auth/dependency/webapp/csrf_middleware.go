package webapp

import (
	"net/http"

	"github.com/gorilla/csrf"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/core/samesite"
	"github.com/authgear/authgear-server/pkg/httputil"
	"github.com/authgear/authgear-server/pkg/jwkutil"
)

type CSRFMiddleware struct {
	Secret *config.CSRFKeyMaterials
	Config *config.ServerConfig
}

func (m *CSRFMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secure := httputil.GetProto(r, m.Config.TrustProxy) == "https"
		options := []csrf.Option{
			csrf.Path("/"),
			csrf.Secure(secure),
			csrf.CookieName(CSRFCookieName),
		}

		useragent := r.UserAgent()
		if samesite.ShouldSendSameSiteNone(useragent, secure) {
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
