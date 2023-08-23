package authenticationflow

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/intl"
)

type IntlMiddleware struct{}

func (m *IntlMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This middleware only creates the holder of preferred language tags.
		// This enables the holder to be mutated later in other places.
		emptyIntl := []string{}
		ctx := intl.WithPreferredLanguageTags(r.Context(), emptyIntl)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
