package webapp

import (
	"net/http"

	"github.com/gorilla/csrf"
)

type CSRFMiddleware struct {
	Key               string
	UseInsecureCookie bool
}

func (m *CSRFMiddleware) Handle(next http.Handler) http.Handler {
	gorillaCSRF := csrf.Protect(
		[]byte(m.Key),
		csrf.Path("/"),
		csrf.Secure(!m.UseInsecureCookie),
		csrf.SameSite(csrf.SameSiteNoneMode),
		csrf.CookieName(csrfCookieName),
	)
	return gorillaCSRF(next)
}
