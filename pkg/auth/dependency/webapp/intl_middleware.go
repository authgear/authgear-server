package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/intl"
)

func IntlMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tags := PreferredLanguageTagsFromRequest(r)
		ctx := intl.WithPreferredLanguageTags(r.Context(), tags)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func PreferredLanguageTagsFromRequest(r *http.Request) (out []string) {
	acceptLanguage := r.Header.Get("Accept-Language")
	// Intentionally not use r.Form here because it may not be parsed.
	if uiLocales := r.URL.Query().Get("ui_locales"); uiLocales != "" {
		return intl.ParseUILocales(uiLocales)
	}
	return intl.ParseAcceptLanguage(acceptLanguage)
}
