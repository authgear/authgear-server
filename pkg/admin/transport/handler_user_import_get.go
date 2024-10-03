package transport

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/userimport"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureUserImportGetRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("GET").
		WithPathPattern("/_api/admin/users/import/:id")
}

type UserImportJobGetter interface {
	GetJob(ctx context.Context, jobID string) (*userimport.Response, error)
}

type UserImportGetHandler struct {
	AppID       config.AppID
	JSON        JSONResponseWriter
	UserImports UserImportJobGetter
}

func (h *UserImportGetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.handle(w, r)
	if err != nil {
		h.JSON.WriteResponse(w, &api.Response{Error: err})
		return
	}
}

func (h *UserImportGetHandler) handle(w http.ResponseWriter, r *http.Request) error {
	jobID := httproute.GetParam(r, "id")

	resp, err := h.UserImports.GetJob(r.Context(), jobID)
	if err != nil {
		return err
	}

	h.JSON.WriteResponse(w, &api.Response{
		Result: resp,
	})
	return nil
}
