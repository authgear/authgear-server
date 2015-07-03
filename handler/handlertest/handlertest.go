package handlertest

import (
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/oursky/ourd/router"
)

// SingleRouteRouter is a router that only serves with a single handler,
// regardless of the requested action.
type SingleRouteRouter router.Router

// NewSingleRouteRouter creates a SingleRouteRouter, mapping the specified
// handler as the only route.
func NewSingleRouteRouter(handler router.Handler, prepareFunc func(*router.Payload)) *SingleRouteRouter {
	r := router.NewRouter()
	r.Map("", handler, func(p *router.Payload, _ *router.Response) int {
		prepareFunc(p)
		return 200
	})
	return (*SingleRouteRouter)(r)
}

// POST invoke the only route mapped on the SingleRouteRouter.
func (r *SingleRouteRouter) POST(body string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("POST", "", strings.NewReader(body))
	resp := httptest.NewRecorder()

	(*router.Router)(r).ServeHTTP(resp, req)
	return resp
}
