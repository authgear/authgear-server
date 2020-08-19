package upstreamapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/session"
)

type Middleware struct{}

func (m *Middleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionInfo, err := session.NewInfoFromHeaders(r.Header)
		if err != nil {
			panic(err)
		}

		r = r.WithContext(WithSessionInfo(r.Context(), sessionInfo))
		next.ServeHTTP(w, r)
	})
}
