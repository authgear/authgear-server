package api

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureAccountManagementV1IdentificationRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("OPTIONS", "POST").WithPathPattern("/api/v1/account/identification")
}

type AccountManagementV1IdentificationHandler struct{}

func (h *AccountManagementV1IdentificationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}
