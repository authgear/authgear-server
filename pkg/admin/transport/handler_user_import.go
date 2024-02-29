package transport

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureUserImportRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("POST").
		WithPathPattern("/_api/admin/users/import")
}

type UserImportHandler struct{}

func (h *UserImportHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {}
