package slogutil

import (
	"log/slog"
	"net/http"
)

func UserAgentMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := GetContextLogger(ctx)
		logger = logger.With(slog.String("user_agent", r.Header.Get("User-Agent")))
		r = r.WithContext(SetContextLogger(ctx, logger))
		next.ServeHTTP(w, r)
	})
}
