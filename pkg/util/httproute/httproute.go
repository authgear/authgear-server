package httproute

import (
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
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

func (r *Router) Add(route Route, h http.Handler) {
	if route.Middleware != nil {
		h = route.Middleware.Handle(h)
	}
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
	if route.Middleware != nil {
		h = route.Middleware.Handle(h)
	}
	r.router.NotFound = h
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(w, req)
}

func GetParam(r *http.Request, name string) string {
	return httprouter.ParamsFromContext(r.Context()).ByName(name)
}
