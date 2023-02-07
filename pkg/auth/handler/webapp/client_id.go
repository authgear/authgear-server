package webapp

import (
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/clientid"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureClientIDRedirectRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET").
		WithPathPattern("/clients/:clientid/*path")
}

type ClientIDRedirectHandler struct{}

func (h *ClientIDRedirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	clientID := httproute.GetParam(r, "clientid")
	path := httproute.GetParam(r, "path")

	ctx := clientid.WithClientID(r.Context(), clientID)
	r = r.WithContext(ctx)

	url, err := url.Parse(path)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	query := url.Query()
	query.Set("client_id", clientID)
	url.RawQuery = query.Encode()

	http.Redirect(w, r, url.String(), http.StatusFound)
}
