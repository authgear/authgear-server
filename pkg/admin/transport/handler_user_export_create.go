package transport

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/userexport"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureUserExportCreateRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("POST").
		WithPathPattern("/_api/admin/users/export")
}

type UserExportCreateHandler struct {
	JSON JSONResponseWriter
}

func (h *UserExportCreateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.handle(w, r)
	if err != nil {
		h.JSON.WriteResponse(w, &api.Response{Error: err})
		return
	}
}

func (h *UserExportCreateHandler) handle(w http.ResponseWriter, r *http.Request) error {
	// TODO: export users
	h.JSON.WriteResponse(w, &api.Response{
		Result: userexport.Response{
			ID:     "dummy_id",
			Status: "pending",
			Request: &userexport.Request{
				Format: "ndjson",
			},
		},
	})
	return nil
}
