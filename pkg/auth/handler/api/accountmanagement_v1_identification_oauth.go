package api

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureAccountManagementV1IdentificationOAuthRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("OPTIONS", "POST").WithPathPattern("/api/v1/account/identification/oauth")
}

type AccountManagementV1IdentificationOAuthHandler struct {
	JSON JSONResponseWriter
}

func (h *AccountManagementV1IdentificationOAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}
