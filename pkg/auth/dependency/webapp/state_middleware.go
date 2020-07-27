package webapp

import (
	"net/http"
)

type StateMiddlewareStates interface {
	Get(instanceID string) (*State, error)
}

type StateMiddleware struct {
	States StateMiddlewareStates
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
			_, err := m.States.Get(sid)
			if err != nil {
				http.Redirect(w, r, MakeURLWithPathWithoutX(r.URL, "/"), http.StatusFound)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
