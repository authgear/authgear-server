package workflow

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/uiparam"
)

type UIParamMiddleware struct{}

func (m *UIParamMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This middleware only creates the holder of the ui params.
		// This enables the holder to be mutated later in other places.
		emptyState := ""
		emptyUILocales := []string{}
		ctx := uiparam.WithUIParam(r.Context(), emptyState, emptyUILocales)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
