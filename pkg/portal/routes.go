package portal

import (
	"fmt"
	"net/http"
	gohttputil "net/http/httputil"

	"github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func NewRouter(p *deps.RootProvider) *httproute.Router {
	router := httproute.NewRouter()
	router.Add(httproute.Route{
		Methods:     []string{"GET"},
		PathPattern: "/healthz",
	}, http.HandlerFunc(httputil.HealthCheckHandler))

	router.Add(httproute.Route{
		Methods:     []string{"GET"},
		PathPattern: "/api",
	}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO(portal): Install a middleware to take headers from request and put the info in context.
		bytes, err := gohttputil.DumpRequest(r, true)
		if err == nil {
			fmt.Printf("%v\n", string(bytes))
		}

		w.Write([]byte("Hello, World"))
	}))

	return router
}
