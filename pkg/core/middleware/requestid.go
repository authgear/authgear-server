package middleware

import (
	"context"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/server/uuid"
)

// RequestIDMiddleware add random request id to request context
type RequestIDMiddleware struct {}

func (m RequestIDMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.New()
		newContext := context.WithValue(r.Context(), "RequestID", requestID)
		r = r.WithContext(newContext)

		w.Header().Set("X-Skygear-Request-Id", requestID)
		next.ServeHTTP(w, r)
	})
}
