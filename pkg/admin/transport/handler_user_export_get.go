package transport

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/userexport"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureUserExportGetRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("GET").
		WithPathPattern("/_api/admin/users/export/:id")
}

type UserExportGetHandler struct {
	JSON JSONResponseWriter
}

func (h *UserExportGetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.handle(w, r)
	if err != nil {
		h.JSON.WriteResponse(w, &api.Response{Error: err})
		return
	}
}

func (h *UserExportGetHandler) handle(w http.ResponseWriter, r *http.Request) error {
	// TODO: get worker task status by id
	taskID := httproute.GetParam(r, "id")
	h.JSON.WriteResponse(w, &api.Response{
		Result: userexport.Response{
			ID:     taskID,
			Status: "pending",
			Request: &userexport.Request{
				Format: "ndjson",
			},
		},
	})
	return nil
}
