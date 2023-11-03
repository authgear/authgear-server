package uiparam

import (
	"net/http"
)

type Middleware struct{}

func (m *Middleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This middleware only creates the holder of the ui params.
		// This enables the holder to be mutated later in other places.
		var empty T
		ctx := WithUIParam(r.Context(), &empty)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
