package webapp

import (
	"net/http"

	interactionflows "github.com/authgear/authgear-server/pkg/auth/dependency/interaction/flows"
)

type StateMiddlewareStates interface {
	Get(id string) (*interactionflows.State, error)
}

type StateMiddleware struct {
	StateStore StateMiddlewareStates
}

func (m *StateMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		q := r.URL.Query()
		sid := q.Get("x_sid")

		if sid != "" {
			_, err := m.StateStore.Get(sid)
			if err != nil {
				RedirectToPathWithoutX(w, r, "/")
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
