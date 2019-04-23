package middleware

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/uuid"

	coreHttp "github.com/skygeario/skygear-server/pkg/core/http"
)

// RequestIDMiddleware add random request id to request context
type RequestIDMiddleware struct{}

func (m RequestIDMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.New()
		r.Header.Set(coreHttp.HeaderRequestID, requestID)
		w.Header().Set(coreHttp.HeaderRequestID, requestID)
		next.ServeHTTP(w, r)
	})
}
