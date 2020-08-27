package session

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/model"
)

// nolint:golint
type SessionInfoMiddleware struct{}

func (m *SessionInfoMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionInfo, err := model.NewSessionInfoFromHeaders(r.Header)
		if err != nil {
			panic(err)
		}

		r = r.WithContext(WithSessionInfo(r.Context(), sessionInfo))
		next.ServeHTTP(w, r)
	})
}
