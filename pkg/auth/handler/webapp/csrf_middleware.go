package webapp

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gorilla/csrf"
	"github.com/gorilla/securecookie"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	webapp "github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/otelauthgear"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/jwkutil"
	"github.com/authgear/authgear-server/pkg/util/otelutil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var CSRFMiddlewareLogger = slogutil.NewLogger("webapp-csrf-middleware")

type CSRFMiddlewareUIImplementationService interface {
	GetUIImplementation() config.UIImplementation
}

type CSRFMiddleware struct {
	Secret                  *config.CSRFKeyMaterials
	TrustProxy              config.TrustProxy
	Cookies                 CookieManager
	BaseViewModel           *viewmodels.BaseViewModeler
	Renderer                Renderer
	UIImplementationService CSRFMiddlewareUIImplementationService
	EnvironmentConfig       *config.EnvironmentConfig
}

func (m *CSRFMiddleware) Handle(next http.Handler) http.Handler {
	if m.EnvironmentConfig.End2EndCSRFProtectionDisabled {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check using golang net.CrossOriginProtection and record metrics only
		secFetchError := httputil.NewCrossOriginProtection().Check(r)
		if secFetchError != nil {
			otelutil.IntCounterAddOne(
				r.Context(),
				otelauthgear.CounterSecFetchCSRFRequestCount,
				otelauthgear.WithStatusError(),
			)
		} else {
			otelutil.IntCounterAddOne(
				r.Context(),
				otelauthgear.CounterSecFetchCSRFRequestCount,
				otelauthgear.WithStatusOk(),
			)
		}

		proto := httputil.GetProto(r, bool(m.TrustProxy))
		if proto == "http" {
			// By default, gorilla/csrf assumes https
			// Thus it is safe in production environment.
			// We need to explicitly tell it is plaintext in development environment.
			r = csrf.PlaintextHTTPRequest(r)
		}

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

		options = append(options, csrf.ErrorHandler(m.makeUnauthorizedHandler(secFetchError)))

		key, err := m.getSecretKey()
		if err != nil {
			panic("webapp: CSRF key not found")
		}
		gorillaCSRF := csrf.Protect(key, options...)
		h := gorillaCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// When we reach here, the CSRF protection is successful.
			otelutil.IntCounterAddOne(
				r.Context(),
				otelauthgear.CounterCSRFRequestCount,
				otelauthgear.WithStatusOk(),
			)

			// The CSRF protection is successful but Sec-Fetch-CSRF is not.
			// Record unmatched metrics.
			if secFetchError != nil {
				ctx := r.Context()
				logger := CSRFMiddlewareLogger.GetLogger(ctx)
				logger.WithError(secFetchError).Warn(ctx,
					"mismatched csrf protection result",
					slog.String("cookies_based_status", "ok"),
					slog.String("sec_fetch_based_status", "error"))

				otelutil.IntCounterAddOne(
					ctx,
					otelauthgear.CounterSecFetchCSRFUnmatchedCount,
					otelauthgear.WithCookiesBasedStatusOK(),
					otelauthgear.WithSecFetchBasedStatusError(),
				)
			}
			next.ServeHTTP(w, r)
		}))
		h.ServeHTTP(w, r)
	})
}

func (m *CSRFMiddleware) makeUnauthorizedHandler(secFetchError error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.unauthorizedHandler(secFetchError, w, r)
	})
}

func (m *CSRFMiddleware) unauthorizedHandler(secFetchError error, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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

	otelutil.IntCounterAddOne(
		ctx,
		otelauthgear.CounterCSRFRequestCount,
		otelauthgear.WithStatusError(),
		otelauthgear.WithCSRFHasOmitCookie(hasOmitCookie),
		otelauthgear.WithCSRFHasNoneCookie(hasNoneCookie),
		otelauthgear.WithCSRFHasLaxCookie(hasLaxCookie),
		otelauthgear.WithCSRFHasStrictCookie(hasStrictCookie),
		otelauthgear.WithGorillaCSRFFailureReason(func() string {
			var val string
			switch {
			case errors.Is(csrfFailureReason, csrf.ErrNoReferer):
				val = "ErrNoReferer"
			case errors.Is(csrfFailureReason, csrf.ErrBadOrigin):
				val = "ErrBadOrigin"
			case errors.Is(csrfFailureReason, csrf.ErrBadReferer):
				val = "ErrBadReferer"
			case errors.Is(csrfFailureReason, csrf.ErrNoToken):
				val = "ErrNoToken"
			case errors.Is(csrfFailureReason, csrf.ErrBadToken):
				val = "ErrBadToken"
			default:
				val = "unknown"
			}
			return val
		}()),
	)

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

	logger := CSRFMiddlewareLogger.GetLogger(ctx)

	// The CSRF protection is not successful but Sec-Fetch-CSRF is.
	// Record unmatched metrics.
	if secFetchError == nil {
		logger.With(
			slog.String("securecookieError", securecookieError),
			slog.Any("csrfFailureReason", csrfFailureReason),
		).Warn(
			ctx,
			"mismatched csrf protection result",
			slog.String("cookies_based_status", "error"),
			slog.String("sec_fetch_based_status", "ok"),
		)

		otelutil.IntCounterAddOne(
			ctx,
			otelauthgear.CounterSecFetchCSRFUnmatchedCount,
			otelauthgear.WithCookiesBasedStatusError(),
			otelauthgear.WithSecFetchBasedStatusOK(),
		)
	}

	logger.With(
		slog.Bool("hasOmitCookie", hasOmitCookie),
		slog.Bool("hasNoneCookie", hasNoneCookie),
		slog.Bool("hasLaxCookie", hasLaxCookie),
		slog.Bool("hasStrictCookie", hasStrictCookie),
		slog.Int("csrfCookieSizeInBytes", csrfCookieSizeInBytes),
		slog.String("maskedCsrfCookieContent", maskedCsrfCookieContent),
		slog.String("securecookieError", securecookieError),
		slog.Any("csrfFailureReason", csrfFailureReason),
	).Warn(ctx, "CSRF Forbidden")

	uiImpl := m.UIImplementationService.GetUIImplementation()

	data := make(map[string]interface{})
	baseViewModel := m.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	switch uiImpl {
	case config.UIImplementationInteraction:
		http.Error(w, fmt.Sprintf("%v - %v handler/auth/webapp",
			http.StatusText(http.StatusForbidden), csrfFailureReason),
			http.StatusForbidden)
	case config.UIImplementationAuthflowV2:
		fallthrough
	default:
		m.Renderer.RenderHTML(w, r, TemplateCSRFErrorHTML, data)
	}
}

func (m *CSRFMiddleware) getSecretKey() ([]byte, error) {
	key, err := jwkutil.ExtractOctetKey(m.Secret.Set, "")
	if err != nil {
		return nil, err
	}

	return key, nil
}
