package oauth

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureProxyRedirectRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET").
		WithPathPattern("/oauth2/redirect")
}

type ProtocolProxyRedirectHandler interface {
	Validate(redirectURIWithQuery string) error
}

type ProxyRedirectHandler struct {
	ProxyRedirectHandler ProtocolProxyRedirectHandler
}

func (h *ProxyRedirectHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	redirectURI := r.URL.Query().Get("redirect_uri")
	err := h.ProxyRedirectHandler.Validate(redirectURI)

	if err == nil {
		oauth.HTMLRedirect(rw, r, redirectURI)
	} else {
		http.Error(rw, err.Error(), 400)
	}
}
