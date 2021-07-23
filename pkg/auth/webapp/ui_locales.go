package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/intl"
)

// UILocalesCookieDef is a HTTP session cookie.
var UILocalesCookieDef = &httputil.CookieDef{
	NameSuffix: "ui_locales",
	Path:       "/",
	SameSite:   http.SameSiteNoneMode,
}

type UILocalesMiddleware struct {
	Cookies CookieManager
}

func (m *UILocalesMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		uiLocales := q.Get("ui_locales")

		// Persist ui_locales into cookie.
		// So that ui_locales no longer need to be present on the query.
		if uiLocales != "" {
			cookie := m.Cookies.ValueCookie(UILocalesCookieDef, uiLocales)
			httputil.UpdateCookie(w, cookie)
		}

		// Restore ui_locales from cookie
		if uiLocales == "" {
			cookie, err := m.Cookies.GetCookie(r, UILocalesCookieDef)
			if err == nil {
				uiLocales = cookie.Value
			}
		}

		// Restore ui_locales into the request context.
		if uiLocales != "" {
			tags := intl.ParseUILocales(uiLocales)
			ctx := intl.WithPreferredLanguageTags(r.Context(), tags)
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}
