package middleware

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/gateway"
)

// InjectableMiddleware is a pointer to a struct
// with `dependency:` tags.
type InjectableMiddleware interface {
	Handle(next http.Handler) http.Handler
}

// Injecter injects dependencies from Dependency into
// Middleware at request time.
type Injecter struct {
	Middleware InjectableMiddleware
	Dependency gateway.DependencyMap
}

// Handle implements gorilla middleware signature.
func (m Injecter) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inject.DefaultRequestInject(m.Middleware, m.Dependency, r)
		actualHandler := m.Middleware.Handle(next)
		actualHandler.ServeHTTP(w, r)
	})
}
