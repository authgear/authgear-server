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

// InjectableMiddlewareFactory allows creating
// middleware per request.
type InjectableMiddlewareFactory interface {
	NewInjectableMiddleware() InjectableMiddleware
}

// Injecter injects dependencies from Dependency into
// Middleware at request time.
type Injecter struct {
	MiddlewareFactory InjectableMiddlewareFactory
	Dependency        gateway.DependencyMap
}

// Handle implements gorilla middleware signature.
func (m Injecter) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		middleware := m.MiddlewareFactory.NewInjectableMiddleware()
		inject.DefaultRequestInject(middleware, m.Dependency, r)
		actualHandler := middleware.Handle(next)
		actualHandler.ServeHTTP(w, r)
	})
}
