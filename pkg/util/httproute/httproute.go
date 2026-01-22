package httproute

import (
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"

	"github.com/authgear/authgear-server/pkg/util/otelutil"
)

type Middleware interface {
	Handle(http.Handler) http.Handler
}

type MiddlewareFunc func(http.Handler) http.Handler

func (f MiddlewareFunc) Handle(h http.Handler) http.Handler {
	return f(h)
}

func Chain(ms ...Middleware) Middleware {
	return MiddlewareFunc(func(h http.Handler) http.Handler {
		for i := len(ms) - 1; i >= 0; i-- {
			h = ms[i].Handle(h)
		}
		return h
	})
}

type Route struct {
	Methods     []string
	PathPattern string
	Middleware  Middleware
}

func (r Route) WithMethods(methods ...string) Route {
	r.Methods = methods
	return r
}

func (r Route) WithPathPattern(pathPattern string) Route {
	r.PathPattern = pathPattern
	return r
}

func (r Route) PrependPathPattern(pathPattern string) Route {
	newPathPattern := pathPattern
	if strings.HasSuffix(pathPattern, "/") {
		if strings.HasPrefix(r.PathPattern, "/") {
			newPathPattern = newPathPattern[:len(newPathPattern)-1] + r.PathPattern
		} else {
			newPathPattern = newPathPattern + r.PathPattern
		}
	} else {
		if strings.HasPrefix(r.PathPattern, "/") {
			newPathPattern = newPathPattern + r.PathPattern
		} else {
			newPathPattern = newPathPattern + "/" + r.PathPattern
		}
	}
	r.PathPattern = newPathPattern
	return r
}

func (r Route) WithMiddleware(middleware Middleware) Route {
	r.Middleware = middleware
	return r
}

type Router struct {
	router *httprouter.Router
}

func NewRouter() *Router {
	r := httprouter.New()
	r.RedirectTrailingSlash = true
	r.RedirectFixedPath = true
	r.HandleMethodNotAllowed = true
	r.HandleOPTIONS = false
	return &Router{r}
}

func (r *Router) Health(h http.Handler) {
	route := Route{
		Methods:     []string{"GET"},
		PathPattern: "/healthz",
	}
	for _, method := range route.Methods {
		r.router.Handler(method, route.PathPattern, h)
	}
}

func (r *Router) Add(route Route, h http.Handler) {
	middlewares := []Middleware{
		MiddlewareFunc(otelutil.WithOtelContext(route.PathPattern)),
		MiddlewareFunc(otelutil.SetupLabeler),
		MiddlewareFunc(otelutil.WithHTTPRoute(route.PathPattern)),
	}

	if route.Middleware != nil {
		middlewares = append(middlewares, route.Middleware)
	}

	finalMiddleware := Chain(middlewares...)
	h = finalMiddleware.Handle(h)

	for _, method := range route.Methods {
		r.router.Handler(method, route.PathPattern, h)
	}
}

func (r *Router) AddRoutes(h http.Handler, routes ...Route) {
	for _, route := range routes {
		r.Add(route, h)
	}
}

func (r *Router) NotFound(route Route, h http.Handler) {
	middlewares := []Middleware{
		MiddlewareFunc(otelutil.SetupLabeler),
		// In case we migrate to ServeMux pattern,
		// "/" means matches every path.
		// https://pkg.go.dev/net/http#hdr-Patterns-ServeMux
		MiddlewareFunc(otelutil.WithHTTPRoute("/")),
	}

	if route.Middleware != nil {
		middlewares = append(middlewares, route.Middleware)
	}

	finalMiddleware := Chain(middlewares...)
	h = finalMiddleware.Handle(h)

	r.router.NotFound = h
}

func (r *Router) HTTPHandler() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Apply a workaround for https://github.com/golang/go/issues/70834
		// Detect early context canceled, and close the connection with best effort.
		if req.Context().Err() != nil {
			rw.Header().Set("Connection", "close")
		}
		r.router.ServeHTTP(rw, req)
	})
}

func GetParam(r *http.Request, name string) string {
	return httprouter.ParamsFromContext(r.Context()).ByName(name)
}
