package useragentblocklist

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/blocklist"
)

type Middleware struct {
	Blocklist *blocklist.Blocklist
}

func NewMiddleware(blocklist *blocklist.Blocklist) *Middleware {
	return &Middleware{Blocklist: blocklist}
}

func (m *Middleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "User-Agent")

		if m.Blocklist != nil && m.Blocklist.IsBlocked(r.UserAgent()) {
			w.Header().Set("Cache-Control", "no-store")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte("Your User-Agent is not allowed to access this resource"))
			return
		}

		next.ServeHTTP(w, r)
	})
}
