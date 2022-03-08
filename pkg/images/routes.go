package images

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/images/deps"
	"github.com/authgear/authgear-server/pkg/images/handler"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func NewRouter(p *deps.RootProvider) *httproute.Router {
	router := httproute.NewRouter()
	router.Add(httproute.Route{
		Methods:     []string{"GET"},
		PathPattern: "/healthz",
	}, http.HandlerFunc(httputil.HealthCheckHandler))

	rootChain := httproute.Chain(
		p.Middleware(newPanicMiddleware),
		p.Middleware(newSentryMiddleware),
	)

	rootRoute := httproute.Route{Middleware: rootChain}
	router.Add(handler.ConfigureGetRoute(rootRoute), p.Handler(newGetHandler))
	router.Add(handler.ConfigurePostRoute(rootRoute), p.Handler(newPostHandler))

	return router
}
