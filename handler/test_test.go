package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/oursky/ourd/router"
)

type SingleRouteRouter router.Router

func newSingleRouteRouter(handler router.Handler, prepareFunc func(*router.Payload)) *SingleRouteRouter {
	r := router.NewRouter()
	r.Map("", handler, func(p *router.Payload, _ *router.Response) int {
		prepareFunc(p)
		return 200
	})
	return (*SingleRouteRouter)(r)
}

func (r *SingleRouteRouter) POST(body string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("POST", "", strings.NewReader(body))
	resp := httptest.NewRecorder()

	(*router.Router)(r).ServeHTTP(resp, req)
	return resp
}
