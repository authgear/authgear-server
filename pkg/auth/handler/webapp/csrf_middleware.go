package webapp

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/csrf"
	"github.com/gorilla/securecookie"
	"github.com/sirupsen/logrus"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/jwkutil"
	"github.com/authgear/authgear-server/pkg/util/log"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	webapp "github.com/authgear/authgear-server/pkg/auth/webapp"
)

type CSRFMiddlewareLogger struct{ *log.Logger }

func NewCSRFMiddlewareLogger(lf *log.Factory) CSRFMiddlewareLogger {
	return CSRFMiddlewareLogger{lf.New("webapp-csrf-middleware")}
}

type CSRFMiddlewareUIImplementationService interface {
	GetUIImplementation() config.UIImplementation
}

type CSRFMiddleware struct {
	Secret                  *config.CSRFKeyMaterials
	TrustProxy              config.TrustProxy
	Cookies                 CookieManager
	Logger                  CSRFMiddlewareLogger
	BaseViewModel           *viewmodels.BaseViewModeler
	Renderer                Renderer
	UIImplementationService CSRFMiddlewareUIImplementationService
}

func (m *CSRFMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookieThatWouldBeWrittenByOurCookieManager := m.Cookies.ValueCookie(webapp.CSRFCookieDef, "unimportant")
		options := []csrf.Option{
			csrf.FieldName(webapp.CSRFFieldName),
			csrf.CookieName(cookieThatWouldBeWrittenByOurCookieManager.Name),
			csrf.Domain(cookieThatWouldBeWrittenByOurCookieManager.Domain),
			csrf.Path(cookieThatWouldBeWrittenByOurCookieManager.Path),
			csrf.Secure(cookieThatWouldBeWrittenByOurCookieManager.Secure),
			csrf.SameSite(csrf.SameSiteMode(cookieThatWouldBeWrittenByOurCookieManager.SameSite)),
			csrf.MaxAge(cookieThatWouldBeWrittenByOurCookieManager.MaxAge),
		}

		options = append(options, csrf.ErrorHandler(http.HandlerFunc(m.unauthorizedHandler)))

		key, err := m.getSecretKey()
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
	omitCookie, err := m.Cookies.GetCookie(r, webapp.CSRFDebugCookieSameSiteOmitDef)
	hasOmitCookie := (err == nil && omitCookie.Value == "exists")

	noneCookie, err := m.Cookies.GetCookie(r, webapp.CSRFDebugCookieSameSiteNoneDef)
	hasNoneCookie := (err == nil && noneCookie.Value == "exists")

	laxCookie, err := m.Cookies.GetCookie(r, webapp.CSRFDebugCookieSameSiteLaxDef)
	hasLaxCookie := (err == nil && laxCookie.Value == "exists")

	strictCookie, err := m.Cookies.GetCookie(r, webapp.CSRFDebugCookieSameSiteStrictDef)
	hasStrictCookie := (err == nil && strictCookie.Value == "exists")

	cookieThatWouldBeWrittenByOurCookieManager := m.Cookies.ValueCookie(webapp.CSRFCookieDef, "unimportant")
	csrfCookie, _ := m.Cookies.GetCookie(r, webapp.CSRFCookieDef)
	csrfCookieSizeInBytes := 0
	maskedCsrfCookieContent := ""
	securecookieError := ""
	csrfFailureReason := csrf.FailureReason(r)
	if csrfCookie != nil {
		// do not return value but length only for debug.
		csrfCookieSizeInBytes = len([]byte(csrfCookie.Value))
		// securecookie uses URLEncoding
		// See https://github.com/gorilla/securecookie/blob/v1.1.2/securecookie.go#L489
		if data, err := base64.URLEncoding.DecodeString(csrfCookie.Value); err != nil {
			csrfToken := string(data)
			maskedTokenParts := make([]string, 0, 4)
			for i, part := range strings.Split(csrfToken, "|") {
				// token format is date|value|mac
				// ref: https://github.com/gorilla/securecookie/blob/eae3c1840ec4adda88a4af683ad0f60bb690e7c2/securecookie.go#L320C30-L320C44
				// we will mask value and sig
				if i == 0 {
					maskedTokenParts = append(maskedTokenParts, part)
					continue
				}
				maskedTokenParts = append(maskedTokenParts, strings.Repeat("*", len(part)))
			}
			maskedCsrfCookieContent = strings.Join(maskedTokenParts, "|")
		}

		// Ask securecookie to decode it once to obtain the underlying error.
		if key, err := m.getSecretKey(); err == nil {
			// Replicate how securecookie was constructed.
			// See https://github.com/gorilla/csrf/blob/v1.7.2/csrf.go#L175
			sc := securecookie.New(key, nil)
			sc.SetSerializer(securecookie.JSONEncoder{})
			sc.MaxAge(cookieThatWouldBeWrittenByOurCookieManager.MaxAge)

			// Token length is 32.
			// https://github.com/gorilla/csrf/blob/v1.7.2/store.go#L46
			token := make([]byte, 32)
			err = sc.Decode(cookieThatWouldBeWrittenByOurCookieManager.Name, csrfCookie.Value, &token)
			if err != nil {
				securecookieError = err.Error()
			}
		}
	}

	m.Logger.WithFields(logrus.Fields{
		"hasOmitCookie":           hasOmitCookie,
		"hasNoneCookie":           hasNoneCookie,
		"hasLaxCookie":            hasLaxCookie,
		"hasStrictCookie":         hasStrictCookie,
		"csrfCookieSizeInBytes":   csrfCookieSizeInBytes,
		"maskedCsrfCookieContent": maskedCsrfCookieContent,
		"securecookieError":       securecookieError,
		"csrfFailureReason":       csrfFailureReason,
	}).Errorf("CSRF Forbidden: %v", csrfFailureReason)

	uiImpl := m.UIImplementationService.GetUIImplementation()

	data := make(map[string]interface{})
	baseViewModel := m.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	switch uiImpl {
	case config.UIImplementationAuthflowV2:
		m.Renderer.RenderHTML(w, r, TemplateCSRFErrorHTML, data)
	case config.UIImplementationAuthflow:
		fallthrough
	case config.UIImplementationDefault:
		fallthrough
	case config.UIImplementationInteraction:
		fallthrough
	default:
		http.Error(w, fmt.Sprintf("%v - %v handler/auth/webapp",
			http.StatusText(http.StatusForbidden), csrfFailureReason),
			http.StatusForbidden)
	}
}

func (m *CSRFMiddleware) getSecretKey() ([]byte, error) {
	key, err := jwkutil.ExtractOctetKey(m.Secret.Set, "")
	if err != nil {
		return nil, err
	}

	return key, nil
}
