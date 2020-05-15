package webapp

import (
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/skygeario/skygear-server/pkg/core/samesite"
)

type CSRFMiddleware struct {
	Key               string
	UseInsecureCookie bool
}

func (m *CSRFMiddleware) Handle(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		useragent := r.UserAgent()
		options := []csrf.Option{
			csrf.Path("/"),
			csrf.Secure(!m.UseInsecureCookie),
			csrf.CookieName(csrfCookieName),
		}
		if samesite.ShouldSendSameSiteNone(useragent) {
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

		gorillaCSRF := csrf.Protect(
			[]byte(m.Key), options...,
		)
		h := gorillaCSRF(next)
		h.ServeHTTP(w, r)
	})
}
