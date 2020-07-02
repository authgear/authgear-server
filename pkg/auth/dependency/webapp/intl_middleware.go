package webapp

import (
	"net/http"
	"strings"

	"golang.org/x/text/language"

	"github.com/authgear/authgear-server/pkg/core/intl"
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
		acceptLanguage = strings.ReplaceAll(uiLocales, " ", ", ")
	}
	tags, _, err := language.ParseAcceptLanguage(acceptLanguage)
	if err != nil {
		return
	}
	for _, tag := range tags {
		out = append(out, tag.String())
	}
	return
}
