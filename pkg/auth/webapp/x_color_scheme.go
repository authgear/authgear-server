package webapp

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func isValidColorScheme(s string) bool {
	return s == "light" || s == "dark"
}

// ColorSchemeCookieDef is a HTTP session cookie.
var ColorSchemeCookieDef = &httputil.CookieDef{
	NameSuffix:    "x_color_scheme",
	Path:          "/",
	SameSite:      http.SameSiteNoneMode,
	IsNonHostOnly: false,
}

type colorSchemeContextKeyType struct{}

var colorSchemeContextKey = colorSchemeContextKeyType{}

type colorSchemeContext struct {
	ColorScheme string
}

func WithColorScheme(ctx context.Context, colorScheme string) context.Context {
	v, ok := ctx.Value(colorSchemeContextKey).(*colorSchemeContext)
	if ok {
		v.ColorScheme = colorScheme
		return ctx
	}

	return context.WithValue(ctx, colorSchemeContextKey, &colorSchemeContext{
		ColorScheme: colorScheme,
	})
}

func GetColorScheme(ctx context.Context) string {
	v, ok := ctx.Value(colorSchemeContextKey).(*colorSchemeContext)
	if !ok {
		return ""
	}
	return v.ColorScheme
}

type ColorSchemeMiddleware struct {
	Cookies CookieManager
}

func (m *ColorSchemeMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		colorScheme := q.Get("x_color_scheme")

		if isValidColorScheme(colorScheme) {
			// Persist to cookie.
			cookie := m.Cookies.ValueCookie(ColorSchemeCookieDef, colorScheme)
			httputil.UpdateCookie(w, cookie)
		}

		// Restore from cookie.
		if colorScheme == "" {
			cookie, err := m.Cookies.GetCookie(r, ColorSchemeCookieDef)
			if err == nil {
				colorScheme = cookie.Value
			}
		}

		// Restore into request context.
		if isValidColorScheme(colorScheme) {
			ctx := WithColorScheme(r.Context(), colorScheme)
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}
