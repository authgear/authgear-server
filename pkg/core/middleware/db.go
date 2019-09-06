package middleware

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/db"
)

// DBMiddleware setup DB context in request
type DBMiddleware struct {
	db.Pool
}

func (m DBMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = db.InitRequestDBContext(r, m.Pool)
		next.ServeHTTP(w, r)
	})
}
