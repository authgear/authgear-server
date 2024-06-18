package webapp

import (
	"fmt"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/sirupsen/logrus"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/jwkutil"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type CSRFMiddlewareLogger struct{ *log.Logger }

func NewCSRFMiddlewareLogger(lf *log.Factory) CSRFMiddlewareLogger {
	return CSRFMiddlewareLogger{lf.New("webapp-csrf-middleware")}
}

type CSRFMiddleware struct {
	Secret     *config.CSRFKeyMaterials
	CookieDef  CSRFCookieDef
	TrustProxy config.TrustProxy
	Cookies    CookieManager
	Logger     CSRFMiddlewareLogger
}

func (m *CSRFMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secure := httputil.GetProto(r, bool(m.TrustProxy)) == "https"
		options := []csrf.Option{
			csrf.FieldName(CSRFFieldName),
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

		options = append(options, csrf.ErrorHandler(http.HandlerFunc(m.unauthorizedHandler)))

		key, err := jwkutil.ExtractOctetKey(m.Secret.Set, "")
		if err != nil {
			panic("webapp: CSRF key not found")
		}
		gorillaCSRF := csrf.Protect(key, options...)
		h := gorillaCSRF(next)
		h.ServeHTTP(w, r)
	})
}

func (m *CSRFMiddleware) unauthorizedHandler(w http.ResponseWriter, r *http.Request) {
	// Check debug cookies and inject info for reporting
	omitCookie, err := m.Cookies.GetCookie(r, CSRFDebugCookieSameSiteOmitDef)
	hasOmitCookie := (err == nil && omitCookie.Value == "exists")

	noneCookie, err := m.Cookies.GetCookie(r, CSRFDebugCookieSameSiteNoneDef)
	hasNoneCookie := (err == nil && noneCookie.Value == "exists")

	laxCookie, err := m.Cookies.GetCookie(r, CSRFDebugCookieSameSiteLaxDef)
	hasLaxCookie := (err == nil && laxCookie.Value == "exists")

	strictCookie, err := m.Cookies.GetCookie(r, CSRFDebugCookieSameSiteStrictDef)
	hasStrictCookie := (err == nil && strictCookie.Value == "exists")

	m.Logger.WithFields(logrus.Fields{
		"hasOmitCookie":   hasOmitCookie,
		"hasNoneCookie":   hasNoneCookie,
		"hasLaxCookie":    hasLaxCookie,
		"hasStrictCookie": hasStrictCookie,
	}).Errorf("CSRF Forbidden: %s", csrf.FailureReason(r))

	// TODO: beautify error page ui
	http.Error(w, fmt.Sprintf("%s - %s",
		http.StatusText(http.StatusForbidden), csrf.FailureReason(r)),
		http.StatusForbidden)
}
