package webapp

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type StateMiddleware struct {
	StateStore StateStore
}

func GetPathComponents(u *url.URL) (out []string) {
	parts := strings.Split(u.EscapedPath(), "/")
	for _, part := range parts {
		if part != "" {
			out = append(out, fmt.Sprintf("/%s", part))
		}
	}
	return
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
				RedirectToPathWithQueryPreserved(w, r, "/")
				return
			}
		}

		// Redirect to / if sid is supposed to be there.
		if sid == "" {
			components := GetPathComponents(r.URL)
			if len(components) > 1 {
				RedirectToPathWithQueryPreserved(w, r, "/")
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
