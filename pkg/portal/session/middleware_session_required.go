package session

import (
	"net/http"
)

// nolint:golint
type SessionRequiredMiddleware struct{}

func (m *SessionRequiredMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionInfo := GetValidSessionInfo(r.Context())
		if sessionInfo == nil {
			http.Error(w, "Session Required", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
